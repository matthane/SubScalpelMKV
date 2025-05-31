package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/devfacet/gocmd/v3"
)

const (
	ErrCodeSuccess = 0
	ErrCodeFailure = 1
)

type MKVTrackProperties struct {
	CodecId              string  `json:"codec_id"`
	TrackName            string  `json:"track_name"`
	Encoding             string  `json:"encoding"`
	Language             string  `json:"language"`
	Number               int     `json:"number"`
	Forced               bool    `json:"forced_track"`
	Default              bool    `json:"default_track"`
	Enabled              bool    `json:"enabled_track"`
	TextSubtitles        bool    `json:"text_subtitles"`
	NumberOfIndexEntries int     `json:"num_index_entries"`
	Duration             string  `json:"tag_duration"`
	UId                  big.Int `json:"uid"`
}

type MKVTrack struct {
	Codec      string             `json:"codec"`
	Id         int                `json:"id"`
	Type       string             `json:"type"`
	Properties MKVTrackProperties `json:"properties"`
}

type MKVContainer struct {
	Type string `json:"type"`
}

type MKVInfo struct {
	Tracks    []MKVTrack   `json:"tracks"`
	Container MKVContainer `json:"container"`
}

var subtitleExtensionByCodec = map[string]string{
	"S_TEXT/UTF8": "srt",
	"S_TEXT/ASS":  "ass",
	"S_HDMV/PGS":  "sup",
}

// Language code mapping from ISO 639-1 (2-letter) to ISO 639-2 (3-letter)
var languageCodeMapping = map[string]string{
	"en": "eng", // English
	"es": "spa", // Spanish
	"fr": "fre", // French
	"de": "ger", // German
	"it": "ita", // Italian
	"pt": "por", // Portuguese
	"ru": "rus", // Russian
	"ja": "jpn", // Japanese
	"ko": "kor", // Korean
	"zh": "chi", // Chinese
	"ar": "ara", // Arabic
	"hi": "hin", // Hindi
	"nl": "dut", // Dutch
	"sv": "swe", // Swedish
	"no": "nor", // Norwegian
	"da": "dan", // Danish
	"fi": "fin", // Finnish
	"pl": "pol", // Polish
	"cs": "cze", // Czech
	"hu": "hun", // Hungarian
	"tr": "tur", // Turkish
	"he": "heb", // Hebrew
	"th": "tha", // Thai
	"vi": "vie", // Vietnamese
}

// matchesLanguageFilter checks if a track language matches the specified filter
// Supports both 2-letter (ISO 639-1) and 3-letter (ISO 639-2) language codes
func matchesLanguageFilter(trackLanguage, filterLanguage string) bool {
	if filterLanguage == "" {
		return true // No filter specified, match all
	}

	// Direct match
	if strings.EqualFold(trackLanguage, filterLanguage) {
		return true
	}

	// Check if filter is 2-letter code and track uses 3-letter code
	if len(filterLanguage) == 2 {
		if mappedCode, exists := languageCodeMapping[strings.ToLower(filterLanguage)]; exists {
			return strings.EqualFold(trackLanguage, mappedCode)
		}
	}

	// Check if filter is 3-letter code and track uses 2-letter code
	if len(filterLanguage) == 3 {
		for twoLetter, threeLetter := range languageCodeMapping {
			if strings.EqualFold(filterLanguage, threeLetter) {
				return strings.EqualFold(trackLanguage, twoLetter)
			}
		}
	}

	return false
}

func isMKVFile(inputFileName string) bool {
	lower := strings.ToLower(inputFileName)
	return strings.HasSuffix(lower, ".mkv") || strings.HasSuffix(lower, ".mks")
}

func buildSubtitlesFileName(inputFileName string, track MKVTrack) string {
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
	outFileName = fmt.Sprintf("%s.%s", outFileName, subtitleExtensionByCodec[track.Properties.CodecId])
	outFileName = path.Join(baseDir, outFileName)
	return outFileName
}

func extractSubtitles(inputFileName string, track MKVTrack, outFileName string) error {
	cmd := exec.Command(
		"mkvextract",
		fmt.Sprintf("%v", inputFileName),
		"tracks",
		fmt.Sprintf("%d:%v", track.Id, outFileName),
	)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		fmt.Printf("Error extracting track %d: %v\n", track.Id, cmdErr)
		fmt.Println(string(output))
		return cmdErr
	}
	fmt.Printf("  ✓ Extracted track %d (%s) -> %s\n", track.Properties.Number, track.Properties.Language, outFileName)
	return nil
}

// showProgressBar displays a progress bar based on percentage
func showProgressBar(percentage int) {
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

// parseProgressLine extracts percentage from mkvmerge progress output
func parseProgressLine(line string) (int, bool) {
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

// createSubtitlesMKS creates a .mks file containing only selected subtitle tracks from the input MKV file
func createSubtitlesMKS(inputFileName string, selection TrackSelection) (string, error) {
	// Create temporary .mks file path
	dir := path.Dir(inputFileName)
	baseName := strings.TrimSuffix(path.Base(inputFileName), path.Ext(inputFileName))
	mksFileName := path.Join(dir, baseName+".subtitles.mks")

	fmt.Println("Step 1: Creating temporary subtitle file...")

	// First, get track information from the original file to determine which tracks to include
	originalMkvInfo, err := getTrackInfo(inputFileName)
	if err != nil {
		return "", fmt.Errorf("failed to analyze original file: %v", err)
	}

	// Build list of subtitle track IDs that match the selection criteria
	var selectedTrackIDs []string
	for _, track := range originalMkvInfo.Tracks {
		if track.Type == "subtitles" {
			// Check if track matches the selection criteria
			if matchesTrackSelection(track, selection) {
				selectedTrackIDs = append(selectedTrackIDs, strconv.Itoa(track.Id))
			}
		}
	}

	// If no tracks match the filter, return an error
	if len(selectedTrackIDs) == 0 {
		return "", fmt.Errorf("no subtitle tracks match the specified selection criteria")
	}

	// Build mkvmerge command with track selection
	args := []string{
		"--gui-mode",
		"-o", mksFileName,
		"--no-video",
		"--no-audio",
		"--no-chapters",
		"--no-attachments",
		"--no-global-tags",
		"--no-track-tags",
	}

	// Add subtitle track selection - only include matching tracks
	if len(selection.LanguageCodes) > 0 || len(selection.TrackNumbers) > 0 {
		subtitleTracks := strings.Join(selectedTrackIDs, ",")
		args = append(args, "--subtitle-tracks", subtitleTracks)
		fmt.Printf("  Including subtitle tracks: %s\n", subtitleTracks)
	}

	args = append(args, inputFileName)
	cmd := exec.Command("mkvmerge", args...)

	// Set up pipe to capture stdout for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mkvmerge: %v", err)
	}

	// Hide cursor for cleaner progress display
	fmt.Print("\033[?25l")

	// Show initial status while mkvmerge initializes
	fmt.Print("Muxing subtitle tracks... [initializing...]")

	// Monitor stdout for progress information
	progressStarted := false
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line contains progress information
		if percentage, isProgress := parseProgressLine(line); isProgress {
			// Only start showing progress bar when we get non-zero progress
			if percentage > 0 && !progressStarted {
				// Clear the initializing message on first real progress update
				fmt.Print("\r\033[K")
				progressStarted = true
			}

			// Only show progress bar if we've started (non-zero progress detected)
			if progressStarted {
				showProgressBar(percentage)
			}
		}
	}

	// Wait for the command to complete
	cmdErr := cmd.Wait()

	// Show cursor again
	fmt.Print("\033[?25h")

	if cmdErr != nil {
		// Clear the progress line before showing error
		fmt.Print("\r\033[K")
		fmt.Printf("Error creating temporary subtitle file: %v\n", cmdErr)
		return "", cmdErr
	}

	return mksFileName, nil
}

// cleanupTempFile removes the temporary .mks file
func cleanupTempFile(fileName string) {
	if fileName != "" {
		if err := os.Remove(fileName); err != nil {
			// Silently ignore cleanup errors - not critical for user
		}
	}
}

// getTrackInfo gets track information from an MKV file using mkvmerge -J
func getTrackInfo(inputFileName string) (*MKVInfo, error) {
	out, cmdErr := exec.Command("mkvmerge", "-J", inputFileName).Output()
	if cmdErr != nil {
		return nil, fmt.Errorf("error analyzing tracks: %v", cmdErr)
	}

	var mkvInfo MKVInfo
	jsonErr := json.Unmarshal(out, &mkvInfo)
	if jsonErr != nil {
		return nil, fmt.Errorf("error parsing track information: %v", jsonErr)
	}

	if !(strings.ToLower(strings.TrimSpace(mkvInfo.Container.Type)) == "matroska") {
		return nil, errors.New("file is not a valid Matroska container")
	}

	return &mkvInfo, nil
}

// displaySubtitleTracks shows available subtitle tracks to the user
func displaySubtitleTracks(mkvInfo *MKVInfo) {
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
			if ext, exists := subtitleExtensionByCodec[track.Properties.CodecId]; exists {
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

// askUserConfirmation asks the user if they want to extract all tracks
func askUserConfirmation() bool {
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

// askTrackSelection asks the user to enter language codes and/or track numbers for selective extraction
func askTrackSelection() string {
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

// parseLanguageCodes parses comma-separated language codes and validates them
func parseLanguageCodes(input string) []string {
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
			_, isValid = languageCodeMapping[strings.ToLower(code)]
		} else if len(code) == 3 {
			for _, threeLetter := range languageCodeMapping {
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

// TrackSelection represents the user's track selection criteria
type TrackSelection struct {
	LanguageCodes []string
	TrackNumbers  []int
}

// parseTrackSelection parses comma-separated language codes and track numbers
func parseTrackSelection(input string) TrackSelection {
	selection := TrackSelection{
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
			_, isValidLanguage = languageCodeMapping[strings.ToLower(item)]
		} else if len(item) == 3 {
			for _, threeLetter := range languageCodeMapping {
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

// matchesTrackSelection checks if a track matches the user's selection criteria
func matchesTrackSelection(track MKVTrack, selection TrackSelection) bool {
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
		if matchesLanguageFilter(track.Properties.Language, langCode) {
			return true
		}
	}

	return false
}

// matchesAnyLanguageFilter checks if a track language matches any of the specified filters
func matchesAnyLanguageFilter(trackLanguage string, languageFilters []string) bool {
	if len(languageFilters) == 0 {
		return true // No filters specified, match all
	}

	for _, filter := range languageFilters {
		if matchesLanguageFilter(trackLanguage, filter) {
			return true
		}
	}

	return false
}

// processFile handles the actual subtitle extraction logic
func processFile(inputFileName, languageFilter string, showFilterMessage bool) error {
	// Parse track selection (language codes and/or track numbers)
	var selection TrackSelection
	if languageFilter != "" {
		selection = parseTrackSelection(languageFilter)
		if showFilterMessage {
			if len(selection.LanguageCodes) > 0 && len(selection.TrackNumbers) > 0 {
				fmt.Printf("Track filter: languages %v and track numbers %v (only muxing and extracting matching tracks)\n",
					selection.LanguageCodes, selection.TrackNumbers)
			} else if len(selection.LanguageCodes) > 0 {
				fmt.Printf("Language filter: %v (only muxing and extracting matching tracks)\n", selection.LanguageCodes)
			} else {
				fmt.Printf("Track number filter: %v (only muxing and extracting matching tracks)\n", selection.TrackNumbers)
			}
		}
	} else if showFilterMessage {
		fmt.Println("No filter - muxing and extracting all subtitle tracks")
	}

	if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
		fmt.Printf("Error: File does not exist or is a directory: %s\n", inputFileName)
		return statErr
	}
	if !isMKVFile(inputFileName) {
		fmt.Printf("Error: File is not an MKV file: %s\n", inputFileName)
		return errors.New("file is not an MKV file")
	}

	// Step 1: Create .mks file with only selected subtitle tracks for optimized processing
	mksFileName, mksErr := createSubtitlesMKS(inputFileName, selection)
	if mksErr != nil {
		return mksErr
	}
	// Ensure cleanup of temporary .mks file
	defer cleanupTempFile(mksFileName)

	// Step 2: Get track information from the .mks file
	fmt.Println("Step 2: Analyzing subtitle tracks...")
	out, cmdErr := exec.Command("mkvmerge", "-J", mksFileName).Output()
	if cmdErr != nil {
		fmt.Printf("Error analyzing subtitle tracks: %v\n", cmdErr)
		return cmdErr
	}
	var mkvInfo MKVInfo
	jsonErr := json.Unmarshal(out, &mkvInfo)
	if jsonErr != nil {
		fmt.Printf("Error parsing track information: %v\n", jsonErr)
		return jsonErr
	}
	if !(strings.ToLower(strings.TrimSpace(mkvInfo.Container.Type)) == "matroska") {
		fmt.Printf("Error: File is not a valid Matroska container\n")
		return errors.New("file is not a valid Matroska container")
	}

	// Step 3: Extract subtitles from the .mks file
	fmt.Println("Step 3: Extracting subtitle tracks...")
	extractedCount := 0
	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			// All tracks in the MKS file are already filtered, so extract them all
			outFileName := buildSubtitlesFileName(inputFileName, track)
			// Use .mks file instead of original video file for extraction
			extractSubsErr := extractSubtitles(mksFileName, track, outFileName)
			if extractSubsErr != nil {
				fmt.Printf("Error extracting subtitles: %v\n", extractSubsErr)
				return extractSubsErr
			}
			extractedCount++
		}
	}

	fmt.Println()
	if extractedCount == 0 {
		fmt.Println("No subtitle tracks found or no tracks matched the language filter")
	} else {
		fmt.Printf("✓ Successfully extracted %d subtitle track(s)\n", extractedCount)
	}

	return nil
}

// showHelp displays the help message
func showHelp() {
	fmt.Println("\nUsage:")
	fmt.Println("  subscalpelmkv [OPTIONS] <file>")
	fmt.Println("  subscalpelmkv -x <file> [selection options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -x, --extract <file>       Extract subtitles from MKV file")
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

func main() {
	fmt.Println("SubScalpelMKV")
	fmt.Println("=============")

	// Check for help flags first, before any other processing
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			showHelp()
			os.Exit(ErrCodeSuccess)
		}
	}

	// Check if file was dragged and dropped (arguments passed directly)
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Drag and drop mode - reconstruct the file path from all arguments
		inputFileName := strings.Join(args, " ")

		fmt.Printf("Processing file: %s\n", inputFileName)

		// Validate file exists and is MKV
		if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
			fmt.Printf("Error: File does not exist or is a directory: %s\n", inputFileName)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}
		if !isMKVFile(inputFileName) {
			fmt.Printf("Error: File is not an MKV file: %s\n", inputFileName)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		// Get track information to show available subtitle tracks
		fmt.Println("Analyzing file...")
		mkvInfo, err := getTrackInfo(inputFileName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		// Display available subtitle tracks
		displaySubtitleTracks(mkvInfo)

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
			os.Exit(ErrCodeSuccess)
		}

		// Ask user if they want to extract all tracks
		extractAll := askUserConfirmation()

		var languageFilter string
		if !extractAll {
			// Ask for specific track selection (language codes and/or track numbers)
			selectionInput := askTrackSelection()
			selection := parseTrackSelection(selectionInput)

			if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 {
				fmt.Println("No valid language codes or track numbers provided. Exiting.")
				fmt.Println("Press Enter to exit...")
				fmt.Scanln()
				os.Exit(ErrCodeSuccess)
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
			fmt.Println("\nExtracting all subtitle tracks...\n")
		}
		err = processFile(inputFileName, languageFilter, false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		os.Exit(ErrCodeSuccess)
	}

	// Command line flag mode (original behavior)
	flags := struct {
		Extract   string `short:"x" long:"extract" description:"Extract subtitles from MKV file" required:"true"`
		Language  string `short:"l" long:"language" description:"Language codes to filter subtitle tracks (e.g., 'eng', 'spa', 'fre'). Use comma-separated values for multiple languages. If not specified, all subtitle tracks will be extracted"`
		Tracks    string `short:"t" long:"tracks" description:"Specific track numbers to extract (e.g., '3,5,7'). Use comma-separated values for multiple tracks"`
		Selection string `short:"s" long:"selection" description:"Mixed selection of language codes and track numbers (e.g., 'eng,3,spa,7'). Combines language and track filtering"`
	}{}

	_, extractHandleFlagErr := gocmd.HandleFlag("Extract", func(cmd *gocmd.Cmd, args []string) error {
		var inputFileName = flags.Extract

		// Build selection filter from command line arguments
		var selectionFilter string

		// Priority: selection > tracks > language
		if flags.Selection != "" {
			selectionFilter = flags.Selection
		} else if flags.Tracks != "" {
			if flags.Language != "" {
				// Combine tracks and language
				selectionFilter = flags.Language + "," + flags.Tracks
			} else {
				selectionFilter = flags.Tracks
			}
		} else {
			selectionFilter = flags.Language
		}

		return processFile(inputFileName, selectionFilter, true)
	})

	if extractHandleFlagErr != nil {
		fmt.Printf("Error handling command flags: %v\n", extractHandleFlagErr)
		os.Exit(ErrCodeFailure)
	}

	_, cmdErr := gocmd.New(gocmd.Options{
		Name:        "subscalpelmkv",
		Description: "SubScalpelMKV - Extract subtitle tracks from MKV files quickly and precisely like a scalpel blade. Supports drag-and-drop: simply drag an MKV file onto the executable.",
		Version:     "1.0.0",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})

	if cmdErr != nil {
		fmt.Printf("Error creating command: %v\n", cmdErr)
		return
	}

	os.Exit(ErrCodeSuccess)
}
