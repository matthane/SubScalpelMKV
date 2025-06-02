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
	// Determine output directory
	var outputDir string
	if config.OutputDir != "" {
		outputDir = config.OutputDir
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

	// Build filename using template
	fileName := BuildFileNameFromTemplate(inputFileName, track, config.Template)

	return filepath.Join(outputDir, fileName)
}

// BuildFileNameFromTemplate builds a filename using a template with placeholders
func BuildFileNameFromTemplate(inputFileName string, track model.MKVTrack, template string) string {
	if template == "" {
		template = model.DefaultOutputTemplate
	}

	// Extract components from input filename
	fileName := filepath.Base(inputFileName)
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)

	// Get subtitle extension
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

	// Build replacement map
	replacements := map[string]string{
		"{basename}":  baseName,
		"{language}":  track.Properties.Language,
		"{trackno}":   trackNo,
		"{trackname}": track.Properties.TrackName,
		"{forced}":    "",
		"{default}":   "",
		"{extension}": subtitleExt,
	}

	// Handle conditional flags
	if track.Properties.Forced {
		replacements["{forced}"] = "forced"
	}
	if track.Properties.Default {
		replacements["{default}"] = "default"
	}

	// Apply replacements
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Clean up multiple consecutive dots and trailing dots
	result = cleanupFileName(result)

	return result
}

// cleanupFileName removes empty segments and cleans up the filename
func cleanupFileName(filename string) string {
	// Split by dots and remove empty segments
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
	// If no selection criteria, match all
	if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 {
		return true
	}

	// Check if track number matches (prioritize over language codes)
	for _, trackNum := range selection.TrackNumbers {
		if track.Properties.Number == trackNum {
			return true
		}
	}

	// Check if language matches
	for _, langCode := range selection.LanguageCodes {
		if model.MatchesLanguageFilter(track.Properties.Language, langCode) {
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

// ResetProgressBar resets the progress bar for a new operation
func ResetProgressBar() {
	progress.ResetProgressBar()
}

// ParseProgressLine extracts percentage from mkvmerge progress output
func ParseProgressLine(line string) (int, bool) {
	return progress.ParseProgressLine(line)
}
