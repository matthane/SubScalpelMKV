package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/devfacet/gocmd/v3"

	"subscalpelmkv/internal/cli"
	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

const (
	ErrCodeSuccess = 0
	ErrCodeFailure = 1
)

// processFile handles the actual subtitle extraction logic
func processFile(inputFileName, languageFilter string, showFilterMessage bool, outputConfig model.OutputConfig) error {
	// Parse track selection using cli package
	var selection model.TrackSelection
	if languageFilter != "" {
		selection = cli.ParseTrackSelection(languageFilter)
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

	// Validate input file using util package
	if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
		fmt.Printf("Error: File does not exist or is a directory: %s\n", inputFileName)
		return statErr
	}
	if !util.IsMKVFile(inputFileName) {
		fmt.Printf("Error: File is not an MKV file: %s\n", inputFileName)
		return errors.New("file is not an MKV file")
	}

	// Step 0: Get original track information to preserve track numbers
	fmt.Println("Analyzing original file...")
	originalMkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		fmt.Printf("Error analyzing original file: %v\n", err)
		return err
	}

	// Create an ordered list of original tracks that match the selection criteria
	// This preserves the order in which tracks appear in the original file
	var selectedOriginalTracks []model.MKVTrack
	for _, track := range originalMkvInfo.Tracks {
		if track.Type == "subtitles" && util.MatchesTrackSelection(track, selection) {
			selectedOriginalTracks = append(selectedOriginalTracks, track)
		}
	}

	// Step 1: Create .mks file with only selected subtitle tracks using mkv package
	mksFileName, mksErr := mkv.CreateSubtitlesMKS(inputFileName, selection, util.MatchesTrackSelection)
	if mksErr != nil {
		return mksErr
	}
	// Ensure cleanup of temporary .mks file using mkv package
	defer mkv.CleanupTempFile(mksFileName)

	// Step 2: Get track information from the temporary .mks file
	fmt.Println("Step 2: Analyzing subtitle tracks...")
	mkvInfo, err := mkv.GetTrackInfo(mksFileName)
	if err != nil {
		fmt.Printf("Error analyzing subtitle tracks: %v\n", err)
		return err
	}

	// Step 3: Extract subtitles using mkv and util packages
	fmt.Println("Step 3: Extracting subtitle tracks...")
	extractedCount := 0
	mksTrackIndex := 0

	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			// Use the corresponding original track based on order
			// The .mks file should contain tracks in the same order as they were selected
			var originalTrack model.MKVTrack
			if mksTrackIndex < len(selectedOriginalTracks) {
				originalTrack = selectedOriginalTracks[mksTrackIndex]
			} else {
				fmt.Printf("Warning: Track index mismatch, using renumbered track info for track %d\n", track.Id)
				originalTrack = track
			}
			mksTrackIndex++

			// Build output filename using the original track information and output config
			outFileName := util.BuildSubtitlesFileNameWithConfig(inputFileName, originalTrack, outputConfig)
			// Extract subtitles using mkv package (use the .mks file track ID for extraction)
			extractSubsErr := mkv.ExtractSubtitles(mksFileName, track, outFileName, originalTrack.Properties.Number)
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

func main() {
	fmt.Println("🗡️ SubScalpelMKV")
	fmt.Println("=============")

	// Parse command-line arguments using gocmd
	args := os.Args[1:]

	// Check for help flags first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			cli.ShowHelp()
			os.Exit(ErrCodeSuccess)
		}
	}

	// Detect execution mode: drag-and-drop vs CLI
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Drag-and-drop mode
		inputFileName := strings.Join(args, " ")

		// Validate file exists and is MKV
		if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
			fmt.Printf("Error: File does not exist or is a directory: %s\n", inputFileName)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}
		if !util.IsMKVFile(inputFileName) {
			fmt.Printf("Error: File is not an MKV file: %s\n", inputFileName)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		// Handle drag-and-drop mode using CLI package with default output config
		defaultOutputConfig := model.OutputConfig{
			OutputDir: "",
			Template:  model.DefaultOutputTemplate,
			CreateDir: false,
		}
		err := cli.HandleDragAndDropModeWithConfig(inputFileName, processFile, defaultOutputConfig)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
		os.Exit(ErrCodeSuccess)
	}

	// CLI mode - set up command-line flags
	flags := struct {
		Extract        string `short:"x" long:"extract" description:"Extract subtitles from MKV file"`
		Info           string `short:"i" long:"info" description:"Display subtitle track information for MKV file"`
		Language       string `short:"l" long:"language" description:"Language codes to filter subtitle tracks (e.g., 'eng', 'spa', 'fre'). Use comma-separated values for multiple languages. If not specified, all subtitle tracks will be extracted"`
		Tracks         string `short:"t" long:"tracks" description:"Specific track numbers to extract (e.g., '3,5,7'). Use comma-separated values for multiple tracks"`
		Selection      string `short:"s" long:"selection" description:"Mixed selection of language codes and track numbers (e.g., 'eng,3,spa,7'). Combines language and track filtering"`
		OutputDir      string `short:"o" long:"output-dir" description:"Output directory for extracted subtitle files. If not specified, uses the same directory as the input file"`
		OutputTemplate string `short:"f" long:"format" description:"Custom filename template with placeholders: {basename}, {language}, {trackno}, {trackname}, {forced}, {default}, {extension}"`
		CreateDir      bool   `short:"c" long:"create-dir" description:"Create output directory if it doesn't exist"`
	}{}

	// Initialize gocmd
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

	// Check which flag was provided and handle accordingly
	if flags.Extract != "" && flags.Info != "" {
		fmt.Println("Error: Cannot use both --extract and --info flags simultaneously")
		os.Exit(ErrCodeFailure)
	}

	if flags.Extract != "" {
		// Handle extract flag
		inputFileName := flags.Extract
		selectionFilter := cli.BuildSelectionFilter(flags.Language, flags.Tracks, flags.Selection)

		// Build output configuration
		outputConfig := model.OutputConfig{
			OutputDir: flags.OutputDir,
			Template:  flags.OutputTemplate,
			CreateDir: flags.CreateDir,
		}

		// Use default template if none specified
		if outputConfig.Template == "" {
			outputConfig.Template = model.DefaultOutputTemplate
		}

		err := processFile(inputFileName, selectionFilter, true, outputConfig)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else if flags.Info != "" {
		// Handle info flag
		inputFileName := flags.Info
		err := cli.ShowFileInfo(inputFileName)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else {
		// No flags provided, show help
		cli.ShowHelp()
		os.Exit(ErrCodeFailure)
	}

	os.Exit(ErrCodeSuccess)
}
