package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/progress"
)

// IsMKVFile checks if the given filename is an MKV file
func IsMKVFile(inputFileName string) bool {
	lower := strings.ToLower(inputFileName)
	return strings.HasSuffix(lower, ".mkv") || strings.HasSuffix(lower, ".mks")
}

// BuildSubtitlesFileName builds the output filename for extracted subtitles
func BuildSubtitlesFileName(inputFileName string, track model.MKVTrack) string {
	// Use default configuration for backward compatibility
	config := model.OutputConfig{
		OutputDir: "",
		Template:  model.DefaultOutputTemplate,
		CreateDir: false,
	}
	return BuildSubtitlesFileNameWithConfig(inputFileName, track, config)
}

// BuildSubtitlesFileNameWithConfig builds the output filename using custom configuration
func BuildSubtitlesFileNameWithConfig(inputFileName string, track model.MKVTrack, config model.OutputConfig) string {
	var outputDir string
	if config.OutputDir != "" {
		// Handle special case for batch mode with -o flag without arguments
		if config.OutputDir == "BATCH_BASENAME_SUBTITLES" {
			baseName := strings.TrimSuffix(filepath.Base(inputFileName), filepath.Ext(inputFileName))
			outputDir = filepath.Join(filepath.Dir(inputFileName), baseName+"-subtitles")
		} else {
			outputDir = config.OutputDir
		}
	} else {
		outputDir = filepath.Dir(inputFileName)
	}

	// Always create output directory if it doesn't exist and a custom output directory is specified
	if config.OutputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Warning: Could not create output directory %s: %v\n", outputDir, err)
			// Fall back to input file directory
			outputDir = filepath.Dir(inputFileName)
		}
	}

	fileName := BuildFileNameFromTemplate(inputFileName, track, config.Template)

	return filepath.Join(outputDir, fileName)
}

// BuildFileNameFromTemplate builds a filename using a template with placeholders
func BuildFileNameFromTemplate(inputFileName string, track model.MKVTrack, template string) string {
	if template == "" {
		template = model.DefaultOutputTemplate
	}

	fileName := filepath.Base(inputFileName)
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)

	subtitleExt := model.SubtitleExtensionByCodec[track.Properties.CodecId]
	if subtitleExt == "" {
		subtitleExt = "srt" // fallback
	}

	// Special handling for S_VOBSUB: ensure we use .sub extension
	// (mkvextract will create both .idx and .sub files automatically)
	if track.Properties.CodecId == "S_VOBSUB" {
		subtitleExt = "sub"
	}

	// Format track number with leading zeros
	trackNo := fmt.Sprintf("%03d", track.Properties.Number)

	replacements := map[string]string{
		"{basename}":  baseName,
		"{language}":  track.Properties.Language,
		"{trackno}":   trackNo,
		"{trackname}": sanitizeFileName(track.Properties.TrackName),
		"{forced}":    "",
		"{default}":   "",
		"{extension}": subtitleExt,
	}

	if track.Properties.Forced {
		replacements["{forced}"] = "forced"
	}
	if track.Properties.Default {
		replacements["{default}"] = "default"
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Clean up multiple consecutive dots and trailing dots
	result = cleanupFileName(result)

	return result
}

// sanitizeFileName removes or replaces characters that are invalid in filenames
func sanitizeFileName(filename string) string {
	if filename == "" {
		return ""
	}
	
	// Replace problematic characters with safe alternatives
	replacements := map[string]string{
		"/": "-",     // Forward slash
		"\\": "-",    // Backslash
		":": "-",     // Colon
		"*": "",      // Asterisk
		"?": "",      // Question mark
		"\"": "",     // Double quote
		"<": "",      // Less than
		">": "",      // Greater than
		"|": "-",     // Pipe
	}
	
	result := filename
	for invalid, replacement := range replacements {
		result = strings.ReplaceAll(result, invalid, replacement)
	}
	
	// Remove leading/trailing spaces and dots
	result = strings.Trim(result, " .")
	
	return result
}

// cleanupFileName removes empty segments and cleans up the filename
func cleanupFileName(filename string) string {
	parts := strings.Split(filename, ".")
	var cleanParts []string

	for _, part := range parts {
		if part != "" {
			cleanParts = append(cleanParts, part)
		}
	}

	return strings.Join(cleanParts, ".")
}

// MatchesTrackSelection checks if a track matches the user's selection criteria
func MatchesTrackSelection(track model.MKVTrack, selection model.TrackSelection) bool {
	// First check if track should be excluded
	if MatchesTrackExclusion(track, selection.Exclusions) {
		return false
	}

	// If no selection criteria, match all (after exclusions)
	if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 && len(selection.FormatFilters) == 0 {
		return true
	}

	// Check if track number matches (prioritize over other criteria)
	for _, trackNum := range selection.TrackNumbers {
		if track.Properties.Number == trackNum {
			return true
		}
	}

	// Check if language matches (additive OR logic)
	for _, langCode := range selection.LanguageCodes {
		if model.MatchesLanguageFilter(track.Properties.Language, langCode) {
			return true
		}
	}

	// Check if format matches (additive OR logic)
	for _, formatFilter := range selection.FormatFilters {
		if model.MatchesFormatFilter(track.Properties.CodecId, formatFilter) {
			return true
		}
	}

	return false
}

// MatchesTrackExclusion checks if a track matches any of the exclusion criteria
func MatchesTrackExclusion(track model.MKVTrack, exclusion model.TrackExclusion) bool {
	// If no exclusion criteria, don't exclude any tracks
	if len(exclusion.LanguageCodes) == 0 && len(exclusion.TrackNumbers) == 0 && len(exclusion.FormatFilters) == 0 {
		return false
	}

	// Check if track number matches exclusion
	for _, trackNum := range exclusion.TrackNumbers {
		if track.Properties.Number == trackNum {
			return true
		}
	}

	// Check if language matches exclusion
	for _, langCode := range exclusion.LanguageCodes {
		if model.MatchesLanguageFilter(track.Properties.Language, langCode) {
			return true
		}
	}

	// Check if format matches exclusion
	for _, formatFilter := range exclusion.FormatFilters {
		if model.MatchesFormatFilter(track.Properties.CodecId, formatFilter) {
			return true
		}
	}

	return false
}

// MatchesAnyLanguageFilter checks if a track language matches any of the specified filters
func MatchesAnyLanguageFilter(trackLanguage string, languageFilters []string) bool {
	if len(languageFilters) == 0 {
		return true // No filters specified, match all
	}

	for _, filter := range languageFilters {
		if model.MatchesLanguageFilter(trackLanguage, filter) {
			return true
		}
	}

	return false
}

// ShowProgressBar displays a progress bar based on percentage
func ShowProgressBar(percentage int) {
	progress.ShowProgressBar(percentage)
}

// UpdateElapsedTime updates only the elapsed time without changing the percentage
func UpdateElapsedTime() {
	progress.UpdateElapsedTime()
}

// ResetProgressBar resets the progress bar for a new operation
func ResetProgressBar() {
	progress.ResetProgressBar()
}

// ParseProgressLine extracts percentage from mkvmerge progress output
func ParseProgressLine(line string) (int, bool) {
	return progress.ParseProgressLine(line)
}

// FindMKVFilesInDirectory recursively finds all MKV files in a directory
func FindMKVFilesInDirectory(dir string) ([]string, error) {
	var mkvFiles []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/directories with errors
		}
		
		if !info.IsDir() && IsMKVFile(path) {
			mkvFiles = append(mkvFiles, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return mkvFiles, nil
}
