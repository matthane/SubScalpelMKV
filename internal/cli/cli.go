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
		format.PrintPromptWithPlaceholder("Extract all tracks? Y/n:", " (press enter for yes)")
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
	format.PrintPromptWithPlaceholder("Selection:", " (press enter to accept all)")

	input, err := reader.ReadString('\n')
	if err != nil {
		format.PrintError(fmt.Sprintf("Error reading input: %v", err))
		return ""
	}

	return strings.TrimSpace(input)
}

// AskTrackExclusion asks the user to enter exclusion criteria for tracks to exclude
func AskTrackExclusion() string {
	reader := bufio.NewReader(os.Stdin)

	format.PrintSubSection("Track Exclusions (Optional)")
	format.PrintInfo("Enter exclusions (comma-separated):")
	format.PrintExample("Language: chi,kor  •  Track ID: 15,17  •  Format: sup,sub  •  Mixed: chi,15,sup")
	format.PrintPromptWithPlaceholder("Exclusions:", " (press enter to skip)")

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
		Exclusions:    model.TrackExclusion{},
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

// ParseTrackExclusion parses comma-separated exclusion criteria (languages, track numbers, formats)
func ParseTrackExclusion(input string) model.TrackExclusion {
	exclusion := model.TrackExclusion{
		LanguageCodes: []string{},
		TrackNumbers:  []int{},
		FormatFilters: []string{},
	}

	if input == "" {
		return exclusion
	}

	items := strings.Split(input, ",")

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Try to parse as track number first
		if trackNum, err := strconv.Atoi(item); err == nil {
			exclusion.TrackNumbers = append(exclusion.TrackNumbers, trackNum)
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
			exclusion.LanguageCodes = append(exclusion.LanguageCodes, item)
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
			exclusion.FormatFilters = append(exclusion.FormatFilters, lowerItem)
		} else {
			format.PrintWarning(fmt.Sprintf("Unknown exclusion language code, format, or invalid track ID '%s' - skipping", item))
		}
	}

	return exclusion
}

// ShowHelp displays the help message
func ShowHelp() {
	format.PrintUsageSection("Usage", `  subscalpelmkv [OPTIONS] <file>
  subscalpelmkv -x <file> [selection options] [output options]
  subscalpelmkv -b <pattern> [selection options] [output options]
  subscalpelmkv -i <file>`)

	format.PrintUsageSection("Selection Options", `  -x, --extract <file>       Extract subtitles from MKV file
	 -b, --batch <pattern>      Extract subtitles from multiple MKV files using glob pattern
	                            (e.g., '*.mkv', 'Season 1/*.mkv', '/path/to/*.mkv')
	 -i, --info <file>          Display subtitle track information
	 -s, --select <selection>   Select subtitle tracks by language codes, track IDs,
	                            and/or subtitle formats. Use comma-separated values.
	                            Language codes: 2-letter (en,es) or 3-letter (eng,spa)
	                            Track IDs: specific track IDs (14,16,18)
	                            Subtitle formats: srt, ass, ssa, sup, sub, vtt, usf, etc.
	                            Mixed: combine all types (e.g., 'eng,14,srt,sup')
	                            If not specified, all subtitle tracks will be extracted
	 -e, --exclude <exclusion>  Exclude subtitle tracks by language codes, track IDs,
	                            and/or subtitle formats. Use comma-separated values.
	                            Same format as --select. Exclusions are applied after
	                            selections, allowing you to exclude specific tracks from
	                            your selection (e.g., 'chi,15,sup')`)

	format.PrintUsageSection("Output Options", `  -o, --output-dir [dir]     Output directory for extracted subtitle files
                             (default: same directory as input file)
                             If -o is used without a directory, creates {basename}-subtitles
                             Output directory will be created if it doesn't exist
  -f, --format <template>    Custom filename template with placeholders:
                             {basename}, {language}, {trackno}, {trackname},
                             {forced}, {default}, {extension}
  -d, --dry-run              Show what would be extracted without performing extraction
  -c, --config               Use default configuration profile
  -p, --profile <name>       Use named configuration profile
  -h, --help                 Show this help message`)

	format.PrintUsageSection("Examples", "")
	format.PrintExample("subscalpelmkv -i video.mkv")
	format.PrintExample("subscalpelmkv -x video.mkv")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng,spa")
	format.PrintExample("subscalpelmkv -x video.mkv -s 14,16")
	format.PrintExample("subscalpelmkv -x video.mkv -s srt,ass")
	format.PrintExample("subscalpelmkv -x video.mkv -s sup")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng,14,srt,sup")
	format.PrintExample("subscalpelmkv -x video.mkv -e chi,kor")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng,spa -e sup")
	format.PrintExample("subscalpelmkv -x video.mkv -e 15,17,sup")
	format.PrintExample("subscalpelmkv -b \"*.mkv\" -s eng")
	format.PrintExample("subscalpelmkv -b \"Season 1/*.mkv\" -s eng,spa")
	format.PrintExample("subscalpelmkv -b \"/path/to/movies/*.mkv\" -o ./subtitles")
	format.PrintExample("subscalpelmkv -x video.mkv -o ./subtitles")
	format.PrintExample("subscalpelmkv -x video.mkv -o")
	format.PrintExample("subscalpelmkv -x video.mkv -f \"{basename}-{language}.{extension}\"")
	format.PrintExample("subscalpelmkv -x video.mkv -s eng --dry-run")
	format.PrintExample("subscalpelmkv -x video.mkv --config")
	format.PrintExample("subscalpelmkv -x video.mkv --profile anime")
	format.PrintExample("subscalpelmkv video.mkv    (drag-and-drop mode)")

	format.PrintUsageSection("Default filename template", `  {basename}.{language}.{trackno}.{trackname}.{forced}.{default}.{extension}`)

	format.PrintUsageSection("Language codes", `  Supports both 2-letter (en, es, fr) and 3-letter (eng, spa, fre) codes`)

	format.PrintUsageSection("Configuration", `  Config files are searched in this order:
  1. ./subscalpelmkv.yaml (current directory)
  2. ~/.config/subscalpelmkv/config.yaml (Linux/macOS)
     %APPDATA%\subscalpelmkv\config.yaml (Windows)
  3. ~/.subscalpelmkv.yaml (home directory)
  
  CLI flags override config values. Use --config for default profile
  or --profile <name> for named profiles.`)

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

			// Get the full language name
			languageName := model.GetLanguageName(track.Properties.Language)

			// For simple SUP tracks without attributes, we need to print codec on second line
			if !track.Properties.Forced && !track.Properties.Default && codecType != "" {
				// Print track info without codec (it will be on second line)
				format.PrintTrackInfoWithLanguageName(
					track.Properties.Number,
					track.Properties.Language,
					languageName,
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
				format.PrintTrackInfoWithLanguageName(
					track.Properties.Number,
					track.Properties.Language,
					languageName,
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
		visibleLen := 2 + len(noTracksMsg)          // "│ " + message
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
		visibleLen := 2 + len(summaryMsg)       // "│ " + message
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
	wrapperFunc := func(inputFileName, languageFilter, exclusionFilter string, showFilterMessage bool, outputConfig model.OutputConfig, dryRun bool) error {
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
func HandleDragAndDropModeWithConfig(inputFileName string, processFileFunc func(string, string, string, bool, model.OutputConfig, bool) error, outputConfig model.OutputConfig) error {
	format.PrintInfo(fmt.Sprintf("Processing file: %s", inputFileName))

	// Get track information to show available subtitle tracks
	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error: %v", err))
		fmt.Println("Press enter to exit...")
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
		fmt.Println("Press enter to exit...")
		fmt.Scanln()
		return nil
	}

	extractAll := AskUserConfirmation()

	// Extract available subtitle track numbers for validation
	var availableTracks []int
	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			availableTracks = append(availableTracks, track.Properties.Number)
		}
	}

	// Use the shared function for processing selection and exclusion
	selectionResult, err := ProcessSelectionAndExclusion(extractAll, availableTracks)
	if err != nil {
		fmt.Println("Press enter to exit...")
		fmt.Scanln()
		return nil
	}

	if selectionResult.Message != "" {
		format.PrintSubSection(selectionResult.Title)
		format.PrintInfo(selectionResult.Message)
	}

	err = processFileFunc(inputFileName, selectionResult.LanguageFilter, selectionResult.ExclusionFilter, false, outputConfig, false)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error: %v", err))
		fmt.Println("Press enter to exit...")
		fmt.Scanln()
		return err
	}

	fmt.Println("Press enter to exit...")
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

	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error analyzing file: %v", err))
		return err
	}

	DisplaySubtitleTracks(mkvInfo)

	return nil
}

// DisplayBatchFiles shows batch file information to the user in the same visual style as subtitle tracks
func DisplayBatchFiles(batchFiles []model.BatchFileInfo) {
	format.PrintSection("Files to Process")

	// Use expanded view as default for batch mode
	for i, fileInfo := range batchFiles {
		if fileInfo.HasError {
			// Display error files differently
			format.BorderColor.Print("│ ")
			format.ErrorColor.Print("✗")
			fmt.Print(" ")
			format.BaseFg.Print(fileInfo.FileName)

			contentLen := 2 + 2 + len(fileInfo.FileName) // "│ " + "✗ " + filename
			padding := format.BoxWidth - contentLen
			if padding > 0 {
				fmt.Print(strings.Repeat(" ", padding))
			}
			format.BorderColor.Println(" │")

			// Error message on second line
			format.BorderColor.Print("│   ")
			format.ErrorColor.Print(fileInfo.ErrorMessage)
			errorLen := 3 + len(fileInfo.ErrorMessage) // "│   " + error
			errorPadding := format.BoxWidth - errorLen - 1
			if errorPadding > 0 {
				fmt.Print(strings.Repeat(" ", errorPadding))
			}
			format.BorderColor.Println(" │")
		} else {
			// Display normal files
			format.BorderColor.Print("│ ")
			format.BaseHighlight.Print("▪")
			fmt.Print(" ")
			format.BaseFg.Print(fileInfo.FileName)

			contentLen := 2 + 2 + len(fileInfo.FileName) // "│ " + "▪ " + filename
			padding := format.BoxWidth - contentLen
			if padding > 0 {
				fmt.Print(strings.Repeat(" ", padding))
			}
			format.BorderColor.Println(" │")

			// Always use expanded view for batch mode
			displayExpandedFileDetails(fileInfo)
		}

		// Add separator between files except for the last one
		if i < len(batchFiles)-1 {
			format.DrawSeparator(format.BoxWidth)
		}
	}

	// Calculate and display summary
	validFiles := 0
	errorFiles := 0
	totalTracks := 0
	languageSet := make(map[string]bool)
	formatSet := make(map[string]bool)

	for _, fileInfo := range batchFiles {
		if fileInfo.HasError {
			errorFiles++
		} else {
			validFiles++
			totalTracks += fileInfo.SubtitleCount

			for _, lang := range fileInfo.LanguageCodes {
				languageSet[lang] = true
			}
			for _, format := range fileInfo.SubtitleFormats {
				formatSet[strings.ToUpper(format)] = true
			}
		}
	}

	if len(batchFiles) > 0 {
		format.DrawSeparator(format.BoxWidth)
	}

	// Display summary
	fileWord := "files"
	if validFiles == 1 {
		fileWord = "file"
	}

	trackWord := "tracks"
	if totalTracks == 1 {
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

	var summaryMsg string
	if errorFiles > 0 {
		summaryMsg = fmt.Sprintf("%d valid %s, %d total %s, %d %s, %d %s • %d errors",
			validFiles, fileWord, totalTracks, trackWord, len(languageSet), languageWord, len(formatSet), formatWord, errorFiles)
	} else {
		summaryMsg = fmt.Sprintf("%d %s, %d total %s, %d %s, %d %s",
			validFiles, fileWord, totalTracks, trackWord, len(languageSet), languageWord, len(formatSet), formatWord)
	}

	visibleLen := 2 + len(summaryMsg) // "│ " + message
	padding := format.BoxWidth - visibleLen
	format.BorderColor.Print("│ ")
	format.InfoColor.Print(summaryMsg)
	if padding > 0 {
		fmt.Print(strings.Repeat(" ", padding))
	}
	format.BorderColor.Println(" │")

	format.DrawBoxBottom(format.BoxWidth)
}

// displayExpandedFileDetails shows all file details across multiple lines
func displayExpandedFileDetails(fileInfo model.BatchFileInfo) {
	// Track count line
	format.BorderColor.Print("│   ")
	trackText := fmt.Sprintf("Tracks: %d", fileInfo.SubtitleCount)
	format.InfoColor.Print(trackText)
	trackLen := 3 + len(trackText)
	trackPadding := format.BoxWidth - trackLen - 1
	if trackPadding > 0 {
		fmt.Print(strings.Repeat(" ", trackPadding))
	}
	format.BorderColor.Println(" │")

	// Languages line (if any)
	if len(fileInfo.LanguageCodes) > 0 {
		// Calculate available width for content
		prefixLen := 3 // "│   "
		suffixLen := 2 // " │"
		availableWidth := format.BoxWidth - prefixLen - suffixLen

		langLabel := "Languages: "
		langLabelLen := len(langLabel)

		// Join all languages
		allLangs := strings.Join(fileInfo.LanguageCodes, ", ")

		// Check if it fits in one line
		if langLabelLen+len(allLangs) <= availableWidth {
			// Single line display
			format.BorderColor.Print("│   ")
			format.BaseDim.Print(langLabel)
			format.BaseAccent.Print(allLangs)

			lineLen := prefixLen + langLabelLen + len(allLangs)
			langPadding := format.BoxWidth - lineLen - 1
			if langPadding > 0 {
				fmt.Print(strings.Repeat(" ", langPadding))
			}
			format.BorderColor.Println(" │")
		} else {
			// Multi-line display with wrapping
			format.BorderColor.Print("│   ")
			format.BaseDim.Print(langLabel)

			// Calculate space remaining on first line
			firstLineSpace := availableWidth - langLabelLen

			// Split languages into lines
			langs := fileInfo.LanguageCodes
			currentLine := ""
			firstLine := true

			for i, lang := range langs {
				// Add comma if not first item
				if i > 0 {
					lang = ", " + lang
				}

				// Check if adding this language would exceed the line width
				testLine := currentLine + lang
				maxWidth := availableWidth - langLabelLen // Continuation lines have less space due to indentation
				if firstLine {
					maxWidth = firstLineSpace
				}

				if len(testLine) > maxWidth && currentLine != "" {
					// Print current line
					if firstLine {
						format.BaseAccent.Print(currentLine)
						padding := format.BoxWidth - prefixLen - langLabelLen - len(currentLine) - 1
						if padding > 0 {
							fmt.Print(strings.Repeat(" ", padding))
						}
						format.BorderColor.Println(" │")
						firstLine = false
					} else {
						format.BorderColor.Print("│   ")
						fmt.Print(strings.Repeat(" ", langLabelLen)) // Indent continuation lines
						format.BaseAccent.Print(currentLine)
						padding := format.BoxWidth - prefixLen - langLabelLen - len(currentLine) - 1
						if padding > 0 {
							fmt.Print(strings.Repeat(" ", padding))
						}
						format.BorderColor.Println(" │")
					}

					// Start new line (remove leading comma and space if present)
					if strings.HasPrefix(lang, ", ") {
						currentLine = lang[2:]
					} else {
						currentLine = lang
					}
				} else {
					currentLine = testLine
				}
			}

			// Print the last line
			if currentLine != "" {
				if firstLine {
					format.BaseAccent.Print(currentLine)
					padding := format.BoxWidth - prefixLen - langLabelLen - len(currentLine) - 1
					if padding > 0 {
						fmt.Print(strings.Repeat(" ", padding))
					}
					format.BorderColor.Println(" │")
				} else {
					format.BorderColor.Print("│   ")
					fmt.Print(strings.Repeat(" ", langLabelLen)) // Indent continuation lines
					format.BaseAccent.Print(currentLine)
					padding := format.BoxWidth - prefixLen - langLabelLen - len(currentLine) - 1
					if padding > 0 {
						fmt.Print(strings.Repeat(" ", padding))
					}
					format.BorderColor.Println(" │")
				}
			}
		}
	}

	// Formats line (if any)
	if len(fileInfo.SubtitleFormats) > 0 {
		// Calculate available width for content
		prefixLen := 3 // "│   "
		suffixLen := 2 // " │"
		availableWidth := format.BoxWidth - prefixLen - suffixLen

		formatLabel := "Formats: "
		formatLabelLen := len(formatLabel)

		// Join all formats
		allFormats := strings.Join(fileInfo.SubtitleFormats, ", ")
		allFormatsUpper := strings.ToUpper(allFormats)

		// Check if it fits in one line
		if formatLabelLen+len(allFormatsUpper) <= availableWidth {
			// Single line display
			format.BorderColor.Print("│   ")
			format.BaseDim.Print(formatLabel)
			format.CodecColor.Print(allFormatsUpper)

			lineLen := prefixLen + formatLabelLen + len(allFormatsUpper)
			formatPadding := format.BoxWidth - lineLen - 1
			if formatPadding > 0 {
				fmt.Print(strings.Repeat(" ", formatPadding))
			}
			format.BorderColor.Println(" │")
		} else {
			// Multi-line display with wrapping
			format.BorderColor.Print("│   ")
			format.BaseDim.Print(formatLabel)

			// Calculate space remaining on first line
			firstLineSpace := availableWidth - formatLabelLen

			// Split formats into lines
			formats := fileInfo.SubtitleFormats
			currentLine := ""
			firstLine := true

			for i, fmtStr := range formats {
				// Add comma if not first item
				fmtUpper := strings.ToUpper(fmtStr)
				if i > 0 {
					fmtUpper = ", " + fmtUpper
				}

				// Check if adding this format would exceed the line width
				testLine := currentLine + fmtUpper
				maxWidth := availableWidth - formatLabelLen // Continuation lines have less space due to indentation
				if firstLine {
					maxWidth = firstLineSpace
				}

				if len(testLine) > maxWidth && currentLine != "" {
					// Print current line
					if firstLine {
						format.CodecColor.Print(currentLine)
						padding := format.BoxWidth - prefixLen - formatLabelLen - len(currentLine) - 1
						if padding > 0 {
							fmt.Print(strings.Repeat(" ", padding))
						}
						format.BorderColor.Println(" │")
						firstLine = false
					} else {
						format.BorderColor.Print("│   ")
						fmt.Print(strings.Repeat(" ", formatLabelLen)) // Indent continuation lines
						format.CodecColor.Print(currentLine)
						padding := format.BoxWidth - prefixLen - formatLabelLen - len(currentLine) - 1
						if padding > 0 {
							fmt.Print(strings.Repeat(" ", padding))
						}
						format.BorderColor.Println(" │")
					}

					// Start new line (remove leading comma and space if present)
					if strings.HasPrefix(fmtUpper, ", ") {
						currentLine = fmtUpper[2:]
					} else {
						currentLine = fmtUpper
					}
				} else {
					currentLine = testLine
				}
			}

			// Print the last line
			if currentLine != "" {
				if firstLine {
					format.CodecColor.Print(currentLine)
					padding := format.BoxWidth - prefixLen - formatLabelLen - len(currentLine) - 1
					if padding > 0 {
						fmt.Print(strings.Repeat(" ", padding))
					}
					format.BorderColor.Println(" │")
				} else {
					format.BorderColor.Print("│   ")
					fmt.Print(strings.Repeat(" ", formatLabelLen)) // Indent continuation lines
					format.CodecColor.Print(currentLine)
					padding := format.BoxWidth - prefixLen - formatLabelLen - len(currentLine) - 1
					if padding > 0 {
						fmt.Print(strings.Repeat(" ", padding))
					}
					format.BorderColor.Println(" │")
				}
			}
		}
	}
}
