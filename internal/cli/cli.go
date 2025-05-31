package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

// AskUserConfirmation asks the user if they want to extract all tracks
func AskUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Extract all tracks? Y/n (default: Y): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
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

		fmt.Println("Please enter 'Y' for yes or 'N' for no.")
	}
}

// AskTrackSelection asks the user to enter language codes and/or track numbers for selective extraction
func AskTrackSelection() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nEnter language codes and/or track numbers separated by commas:")
	fmt.Println("Examples: 'eng,spa,fre' or '3,5,7' or 'eng,3,spa,7'")
	fmt.Println("Language codes: 2-letter (en,es) or 3-letter (eng,spa)")
	fmt.Println("Track numbers: Use the track numbers shown above")
	fmt.Print("Selection: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
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

		// Validate the language code
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
			fmt.Printf("Warning: Unknown language code '%s' - skipping\n", code)
		}
	}

	return validCodes
}

// ParseTrackSelection parses comma-separated language codes and track numbers
func ParseTrackSelection(input string) model.TrackSelection {
	selection := model.TrackSelection{
		LanguageCodes: []string{},
		TrackNumbers:  []int{},
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
		} else {
			fmt.Printf("Warning: Unknown language code or invalid track number '%s' - skipping\n", item)
		}
	}

	return selection
}

// ShowHelp displays the help message
func ShowHelp() {
	fmt.Println("\nUsage:")
	fmt.Println("  subscalpelmkv [OPTIONS] <file>")
	fmt.Println("  subscalpelmkv -x <file> [selection options]")
	fmt.Println("  subscalpelmkv -i <file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -x, --extract <file>       Extract subtitles from MKV file")
	fmt.Println("  -i, --info <file>          Display subtitle track information")
	fmt.Println("  -l, --language <codes>     Language codes to filter subtitle tracks")
	fmt.Println("                             (e.g., 'eng', 'spa,fre'). Use comma-separated")
	fmt.Println("                             values for multiple languages")
	fmt.Println("  -t, --tracks <numbers>     Specific track numbers to extract")
	fmt.Println("                             (e.g., '3', '3,5,7'). Use comma-separated")
	fmt.Println("                             values for multiple tracks")
	fmt.Println("  -s, --selection <mixed>    Mixed selection of language codes and track")
	fmt.Println("                             numbers (e.g., 'eng,3,spa,7')")
	fmt.Println("  -h, --help                 Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  subscalpelmkv -i video.mkv")
	fmt.Println("  subscalpelmkv -x video.mkv")
	fmt.Println("  subscalpelmkv -x video.mkv -l eng")
	fmt.Println("  subscalpelmkv -x video.mkv -l eng,spa")
	fmt.Println("  subscalpelmkv -x video.mkv -t 3,5")
	fmt.Println("  subscalpelmkv -x video.mkv -s eng,3,spa,7")
	fmt.Println("  subscalpelmkv video.mkv    (drag-and-drop mode)")
	fmt.Println()
	fmt.Println("Language codes:")
	fmt.Println("  Supports both 2-letter (en, es, fr) and 3-letter (eng, spa, fre) codes")
	fmt.Println()
	fmt.Println("Drag-and-drop mode:")
	fmt.Println("  Simply drag an MKV file onto the executable for interactive mode")
	fmt.Println("  with track selection options.")
}

// DisplaySubtitleTracks shows available subtitle tracks to the user
func DisplaySubtitleTracks(mkvInfo *model.MKVInfo) {
	fmt.Println("\nAvailable subtitle tracks:")
	fmt.Println("==========================")

	subtitleCount := 0
	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			subtitleCount++
			trackInfo := fmt.Sprintf("Track %d: %s", track.Properties.Number, track.Properties.Language)

			if track.Properties.TrackName != "" {
				trackInfo += fmt.Sprintf(" (%s)", track.Properties.TrackName)
			}

			if track.Properties.Forced {
				trackInfo += " [FORCED]"
			}

			if track.Properties.Default {
				trackInfo += " [DEFAULT]"
			}

			// Show codec type
			codecType := "Unknown"
			if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
				codecType = strings.ToUpper(ext)
			}
			trackInfo += fmt.Sprintf(" [%s]", codecType)

			fmt.Printf("  %s\n", trackInfo)
		}
	}

	if subtitleCount == 0 {
		fmt.Println("  No subtitle tracks found in this file.")
	}
	fmt.Println()
}

// HandleDragAndDropMode handles the interactive drag-and-drop mode
func HandleDragAndDropMode(inputFileName string, processFileFunc func(string, string, bool) error) error {
	fmt.Printf("Processing file: %s\n", inputFileName)

	// Get track information using mkv package to show available subtitle tracks
	fmt.Println("Analyzing file...")
	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return err
	}

	// Display available subtitle tracks
	DisplaySubtitleTracks(mkvInfo)

	// Check if there are any subtitle tracks
	hasSubtitles := false
	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			hasSubtitles = true
			break
		}
	}

	if !hasSubtitles {
		fmt.Println("No subtitle tracks found in this file.")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return nil
	}

	// Ask user if they want to extract all tracks
	extractAll := AskUserConfirmation()

	var languageFilter string
	if !extractAll {
		// Ask for specific track selection
		selectionInput := AskTrackSelection()
		selection := ParseTrackSelection(selectionInput)

		if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 {
			fmt.Println("No valid language codes or track numbers provided. Exiting.")
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			return nil
		}

		// Convert to comma-separated string for processFile function (backward compatibility)
		// Combine language codes and track numbers into a single filter string
		var filterParts []string
		filterParts = append(filterParts, selection.LanguageCodes...)
		for _, trackNum := range selection.TrackNumbers {
			filterParts = append(filterParts, strconv.Itoa(trackNum))
		}
		languageFilter = strings.Join(filterParts, ",")

		if len(selection.LanguageCodes) > 0 && len(selection.TrackNumbers) > 0 {
			fmt.Printf("\nExtracting tracks for languages: %s and track numbers: %v\n\n",
				strings.Join(selection.LanguageCodes, ","), selection.TrackNumbers)
		} else if len(selection.LanguageCodes) > 0 {
			fmt.Printf("\nExtracting tracks for languages: %s\n\n", strings.Join(selection.LanguageCodes, ","))
		} else {
			fmt.Printf("\nExtracting track numbers: %v\n\n", selection.TrackNumbers)
		}
	} else {
		fmt.Println("\nExtracting all subtitle tracks...")
	}

	err = processFileFunc(inputFileName, languageFilter, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return err
	}

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	return nil
}

// BuildSelectionFilter builds a selection filter from command line arguments
func BuildSelectionFilter(language, tracks, selection string) string {
	// Priority: selection > tracks > language
	if selection != "" {
		return selection
	} else if tracks != "" {
		if language != "" {
			// Combine tracks and language
			return language + "," + tracks
		} else {
			return tracks
		}
	} else {
		return language
	}
}

// ShowFileInfo displays subtitle track information for a file without extracting
func ShowFileInfo(inputFileName string) error {
	// Validate input file using util package
	if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
		fmt.Printf("Error: File does not exist or is a directory: %s\n", inputFileName)
		return statErr
	}

	// Check if file is MKV using util package
	if !util.IsMKVFile(inputFileName) {
		fmt.Printf("Error: File is not an MKV file: %s\n", inputFileName)
		return fmt.Errorf("file is not an MKV file")
	}

	fmt.Printf("Analyzing file: %s\n", inputFileName)

	// Get track information using mkv package
	mkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		fmt.Printf("Error analyzing file: %v\n", err)
		return err
	}

	// Display available subtitle tracks
	DisplaySubtitleTracks(mkvInfo)

	return nil
}
