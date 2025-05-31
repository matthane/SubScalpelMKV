package util

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"subscalpelmkv/internal/cli"
	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
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
		outFileName = fmt.Sprintf("%s.%s", outFileName, ".forced")
	}
	outFileName = fmt.Sprintf("%s.%s", outFileName, mkv.SubtitleExtensionByCodec[track.Properties.CodecId])
	outFileName = path.Join(baseDir, outFileName)
	return outFileName
}

// ShowProgressBar displays a progress bar based on percentage
func ShowProgressBar(percentage int) {
	const barWidth = 50
	filled := int(float64(percentage) * float64(barWidth) / 100.0)

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled && percentage < 100 {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	fmt.Printf("\rMuxing subtitle tracks... %s %3d%%", bar, percentage)
	if percentage == 100 {
		fmt.Println(" Complete!")
	}
}

// ParseProgressLine extracts percentage from mkvmerge progress output
func ParseProgressLine(line string) (int, bool) {
	// In GUI mode, progress lines look like: "#GUI#progress 45%"
	if strings.HasPrefix(line, "#GUI#progress ") && strings.HasSuffix(line, "%") {
		percentStr := strings.TrimPrefix(line, "#GUI#progress ")
		percentStr = strings.TrimSuffix(percentStr, "%")
		if percentage, err := strconv.Atoi(strings.TrimSpace(percentStr)); err == nil {
			return percentage, true
		}
	}
	return 0, false
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
		if cli.MatchesLanguageFilter(track.Properties.Language, langCode) {
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
		if cli.MatchesLanguageFilter(trackLanguage, filter) {
			return true
		}
	}

	return false
}
