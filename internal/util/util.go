package util

import (
	"fmt"
	"path"
	"strconv"
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
	baseDir := path.Dir(inputFileName)
	fileName := path.Base(inputFileName)
	extension := path.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)
	trackNo := fmt.Sprintf("%03s", strconv.Itoa(track.Properties.Number))
	outFileName := fmt.Sprintf("%s.%s.%s", baseName, track.Properties.Language, trackNo)
	if track.Properties.TrackName != "" {
		outFileName = fmt.Sprintf("%s.%s", outFileName, track.Properties.TrackName)
	}
	if track.Properties.Forced {
		outFileName = fmt.Sprintf("%s.%s", outFileName, "forced")
	}
	if track.Properties.Default {
		outFileName = fmt.Sprintf("%s.%s", outFileName, "default")
	}
	outFileName = fmt.Sprintf("%s.%s", outFileName, model.SubtitleExtensionByCodec[track.Properties.CodecId])
	outFileName = path.Join(baseDir, outFileName)
	return outFileName
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
