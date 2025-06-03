package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

// AskUserConfirmation asks the user if they want to extract all tracks
func AskUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		format.PrintPrompt("Extract all tracks? Y/n (default: Y): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			format.PrintError(fmt.Sprintf("Error reading input: %v", err))
			continue
		}

		input = strings.TrimSpace(strings.ToLower(input))

		// Default to yes if empty input
		if input == "" || input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		format.PrintWarning("Please enter 'Y' for yes or 'N' for no.")
	}
}

// AskTrackSelection asks the user to enter language codes, track numbers, and/or format filters for selective extraction
func AskTrackSelection() string {
	reader := bufio.NewReader(os.Stdin)

	format.PrintSubSection("Track Selection")
	format.PrintInfo("Enter selection (comma-separated):")
	format.PrintExample("Language: eng,spa,fre  •  Track ID: 14,16,18  •  Format: srt,ass,sup  •  Mixed: eng,14,srt")
	format.PrintPrompt("Selection: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		format.PrintError(fmt.Sprintf("Error reading input: %v", err))
		return ""
	}

	return strings.TrimSpace(input)
}

// ParseLanguageCodes parses comma-separated language codes and validates them
func ParseLanguageCodes(input string) []string {
	if input == "" {
		return []string{}
	}

	codes := strings.Split(input, ",")
	var validCodes []string

	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}

		isValid := false
		if len(code) == 2 {
			_, isValid = model.LanguageCodeMapping[strings.ToLower(code)]
		} else if len(code) == 3 {
			for _, threeLetter := range model.LanguageCodeMapping {
				if strings.EqualFold(code, threeLetter) {
					isValid = true
					break
				}
			}
		}

		if isValid {
			validCodes = append(validCodes, code)
		} else {
			format.PrintWarning(fmt.Sprintf("Unknown language code '%s' - skipping", code))
		}
	}

	return validCodes
}

// ParseTrackSelection parses comma-separated language codes, track numbers, and format filters
func ParseTrackSelection(input string) model.TrackSelection {
	selection := model.TrackSelection{
		LanguageCodes: []string{},
		TrackNumbers:  []int{},
		FormatFilters: []string{},
	}

	if input == "" {
		return selection
	}

	items := strings.Split(input, ",")

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Try to parse as track number first
		if trackNum, err := strconv.Atoi(item); err == nil {
			selection.TrackNumbers = append(selection.TrackNumbers, trackNum)
			continue
		}

		// Try to parse as language code
		isValidLanguage := false
		if len(item) == 2 {
			_, isValidLanguage = model.LanguageCodeMapping[strings.ToLower(item)]
		} else if len(item) == 3 {
			for _, threeLetter := range model.LanguageCodeMapping {
				if strings.EqualFold(item, threeLetter) {
					isValidLanguage = true
					break
				}
			}
		}

		if isValidLanguage {
			selection.LanguageCodes = append(selection.LanguageCodes, item)
			continue
		}

		// Try to parse as subtitle format filter
		isValidFormat := false
		lowerItem := strings.ToLower(item)
		for _, ext := range model.SubtitleExtensionByCodec {
			if lowerItem == ext {
				isValidFormat = true
				break
			}
		}

		if isValidFormat {
			selection.FormatFilters = append(selection.FormatFilters, lowerItem)
		} else {
			format.PrintWarning(fmt.Sprintf("Unknown language code, format, or invalid track ID '%s' - skipping", item))
		}
	}

	return selection
}

// ShowHelp displays the help message
func ShowHelp() {
	format.PrintUsageSection("Usage", `  subscalpelmkv [OPTIONS] <file>
  subscalpelmkv -x <file> [selection options] [output options]
  subscalpelmkv -i <file>

`)

	format.PrintUsageSection("Selection Options", `  -x, --extract <file>       Extract subtitles from MKV file
	 -i, --info <file>          Display subtitle track information
	 -s, --select <selection>   Select subtitle tracks by language codes, track IDs,
	                            and/or subtitle formats. Use comma-separated values.
	                            Language codes: 2-letter (en,es) or 3-letter (eng,spa)
	                            Track IDs: specific track IDs (14,16,18)
	                            Subtitle formats: srt, ass, ssa, sup, sub, vtt, usf, etc.
	                            Mixed: combine all types (e.g., 'eng,14,srt,sup')
	                            If not specified, all subtitle tracks will be extracted

`)

	format.PrintUsageSection("Output Options", `  -o, --output-dir <dir>     Output directory for extracted subtitle files
                             (default: same directory as input file)
                             Output directory will be created if it doesn't exist
  -f, --format <template>    Custom filename template with placeholders:
                             {basename}, {language}, {trackno}, {trackname},
                             {forced}, {default}, {extension}
  -h, --help                 Show this help message

`)

	format.PrintUsageSection("Examples", "")
	format.PrintExample("subscalpelmkv -i video.mkv")
	format.PrintExample("subscalpelmkv -x video.mkv")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng,spa")
	format.PrintExample("subscalpelmkv -x video.mkv -s 14,16")
	format.PrintExample("subscalpelmkv -x video.mkv -s srt,ass")
	format.PrintExample("subscalpelmkv -x video.mkv -s sup")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng,14,srt,sup")
	format.PrintExample("subscalpelmkv -x video.mkv -o ./subtitles")
	format.PrintExample("subscalpelmkv -x video.mkv -f \"{basename}-{language}.{extension}\"")
	format.PrintExample("subscalpelmkv video.mkv    (drag-and-drop mode)")

	format.PrintUsageSection("Default filename template", `  {basename}.{language}.{trackno}.{trackname}.{forced}.{default}.{extension}

`)

	format.PrintUsageSection("Language codes", `  Supports both 2-letter (en, es, fr) and 3-letter (eng, spa, fre) codes

`)

	format.PrintUsageSection("Drag-and-drop mode", `  Simply drag an MKV file onto the executable for interactive mode
  with track selection options.
`)
}

// DisplaySubtitleTracks shows available subtitle tracks to the user
func DisplaySubtitleTracks(mkvInfo *model.MKVInfo) {
	format.PrintSection("Available Subtitle Tracks")

	subtitleCount := 0
	for i, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			subtitleCount++

			codecType := "Unknown"
			if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
				codecType = strings.ToUpper(ext)
			}

			// For simple SUP tracks without attributes, we need to print codec on second line
			if !track.Properties.Forced && !track.Properties.Default && codecType != "" {
				// Print track info without codec (it will be on second line)
				format.PrintTrackInfo(
					track.Properties.Number,
					track.Properties.Language,
					track.Properties.TrackName,
					"", // Empty codec - we'll print it separately
					track.Properties.Forced,
					track.Properties.Default,
				)
				// Print codec on second line
				format.BorderColor.Print("│   ")
				format.CodecColor.Print(codecType)
				// The visible length is 3 (for "   ") + len(codecType)
				visibleLen := 3 + len(codecType)
				padding := format.BoxWidth - visibleLen - 1 // -1 for space before closing border
				if padding > 0 {
					fmt.Print(strings.Repeat(" ", padding))
				}
				format.BorderColor.Println(" │")
			} else {
				// Normal display with attributes
				format.PrintTrackInfo(
					track.Properties.Number,
					track.Properties.Language,
					track.Properties.TrackName,
					codecType,
					track.Properties.Forced,
					track.Properties.Default,
				)
			}
			
			// Add separator between tracks except for the last one
			if i < len(mkvInfo.Tracks)-1 {
				// Check if there are more subtitle tracks after this one
				hasMoreSubtitles := false
				for j := i + 1; j < len(mkvInfo.Tracks); j++ {
					if mkvInfo.Tracks[j].Type == "subtitles" {
						hasMoreSubtitles = true
						break
					}
				}
				if hasMoreSubtitles {
					format.DrawSeparator(format.BoxWidth)
				}
			}
		}
	}

	if subtitleCount == 0 {
		noTracksMsg := "No subtitle tracks found in this file."
		visibleLen := 2 + len(noTracksMsg) // "│ " + message
		padding := format.BoxWidth - visibleLen - 1 // -1 for space before closing border
		format.BorderColor.Print("│ ")
		format.WarningColor.Print(noTracksMsg)
		if padding > 0 {
			fmt.Print(strings.Repeat(" ", padding))
		}
		format.BorderColor.Println(" │")
	} else {
		// Calculate summary statistics
		languageSet := make(map[string]bool)
		formatSet := make(map[string]bool)
		
		for _, track := range mkvInfo.Tracks {
			if track.Type == "subtitles" {
				// Track unique languages
				if track.Properties.Language != "" {
					languageSet[track.Properties.Language] = true
				}
				
				// Track unique formats
				if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
					formatSet[strings.ToUpper(ext)] = true
				}
			}
		}
		
		// Add separator before summary
		if subtitleCount > 0 {
			format.DrawSeparator(format.BoxWidth)
		}
		
		// Display summary
		trackWord := "tracks"
		if subtitleCount == 1 {
			trackWord = "track"
		}
		
		languageWord := "languages"
		if len(languageSet) == 1 {
			languageWord = "language"
		}
		
		formatWord := "formats"
		if len(formatSet) == 1 {
			formatWord = "format"
		}
		
		summaryMsg := fmt.Sprintf("%d total %s, %d %s, %d %s", 
			subtitleCount, trackWord, len(languageSet), languageWord, len(formatSet), formatWord)
		visibleLen := 2 + len(summaryMsg) // "│ " + message
		padding := format.BoxWidth - visibleLen // No -1 needed for proper alignment
		format.BorderColor.Print("│ ")
		format.InfoColor.Print(summaryMsg)
		if padding > 0 {
			fmt.Print(strings.Repeat(" ", padding))
		}
		format.BorderColor.Println(" │")
	}
	
	format.DrawBoxBottom(format.BoxWidth)
}

// HandleDragAndDropMode handles the interactive drag-and-drop mode (backward compatibility)
func HandleDragAndDropMode(inputFileName string, processFileFunc func(string, string, bool) error) error {
	// Create a wrapper function that adds default output config
	wrapperFunc := func(inputFileName, languageFilter string, showFilterMessage bool, outputConfig model.OutputConfig) error {
		return processFileFunc(inputFileName, languageFilter, showFilterMessage)
	}

	defaultOutputConfig := model.OutputConfig{
		OutputDir: "",
		Template:  model.DefaultOutputTemplate,
		CreateDir: false,
	}

	return HandleDragAndDropModeWithConfig(inputFileName, wrapperFunc, defaultOutputConfig)
}

// HandleDragAndDropModeWithConfig handles the interactive drag-and-drop mode with output configuration
func HandleDragAndDropModeWithConfig(inputFileName string, processFileFunc func(string, string, bool, model.OutputConfig) error, outputConfig model.OutputConfig) error {
	format.PrintInfo(fmt.Sprintf("Processing file: %s", inputFileName))

	// Get track information to show available subtitle tracks
	format.PrintInfo("Analyzing file...")
	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error: %v", err))
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return err
	}

	DisplaySubtitleTracks(mkvInfo)

	hasSubtitles := false
	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			hasSubtitles = true
			break
		}
	}

	if !hasSubtitles {
		format.PrintWarning("No subtitle tracks found in this file.")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return nil
	}

	extractAll := AskUserConfirmation()

	var languageFilter string
	if !extractAll {
		selectionInput := AskTrackSelection()
		selection := ParseTrackSelection(selectionInput)

		if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 && len(selection.FormatFilters) == 0 {
			format.PrintWarning("No valid language codes, track IDs, or format filters provided. Exiting.")
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			return nil
		}

		// Convert to comma-separated string for processFile function (backward compatibility)
		// Combine language codes, track numbers, and format filters into a single filter string
		var filterParts []string
		filterParts = append(filterParts, selection.LanguageCodes...)
		for _, trackNum := range selection.TrackNumbers {
			filterParts = append(filterParts, strconv.Itoa(trackNum))
		}
		filterParts = append(filterParts, selection.FormatFilters...)
		languageFilter = strings.Join(filterParts, ",")

		// Build extraction message
		var messageParts []string
		if len(selection.LanguageCodes) > 0 {
			messageParts = append(messageParts, fmt.Sprintf("languages: %s", strings.Join(selection.LanguageCodes, ",")))
		}
		if len(selection.TrackNumbers) > 0 {
			messageParts = append(messageParts, fmt.Sprintf("track IDs: %v", selection.TrackNumbers))
		}
		if len(selection.FormatFilters) > 0 {
			messageParts = append(messageParts, fmt.Sprintf("formats: %s", strings.Join(selection.FormatFilters, ",")))
		}

		if len(messageParts) > 0 {
			format.PrintInfo(fmt.Sprintf("Extracting tracks for %s", strings.Join(messageParts, ", ")))
		}
	} else {
		format.PrintInfo("Extracting all subtitle tracks...")
	}
	fmt.Println()

	err = processFileFunc(inputFileName, languageFilter, false, outputConfig)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error: %v", err))
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return err
	}

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	return nil
}

// BuildSelectionFilter builds a selection filter from command line arguments
func BuildSelectionFilter(input string) string {
	return input
}

// ShowFileInfo displays subtitle track information for a file without extracting
func ShowFileInfo(inputFileName string) error {
	if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
		format.PrintError(fmt.Sprintf("File does not exist or is a directory: %s", inputFileName))
		return statErr
	}

	if !util.IsMKVFile(inputFileName) {
		format.PrintError(fmt.Sprintf("File is not an MKV file: %s", inputFileName))
		return fmt.Errorf("file is not an MKV file")
	}

	format.PrintInfo(fmt.Sprintf("Analyzing file: %s", inputFileName))

	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error analyzing file: %v", err))
		return err
	}

	DisplaySubtitleTracks(mkvInfo)

	return nil
}
