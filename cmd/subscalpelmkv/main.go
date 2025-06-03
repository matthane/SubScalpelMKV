package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/devfacet/gocmd/v3"

	"subscalpelmkv/internal/cli"
	"subscalpelmkv/internal/format"
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
	var selection model.TrackSelection
	if languageFilter != "" {
		selection = cli.ParseTrackSelection(languageFilter)
		if showFilterMessage {
			var filterParts []string
			if len(selection.LanguageCodes) > 0 {
				filterParts = append(filterParts, fmt.Sprintf("languages %v", selection.LanguageCodes))
			}
			if len(selection.TrackNumbers) > 0 {
				filterParts = append(filterParts, fmt.Sprintf("track IDs %v", selection.TrackNumbers))
			}
			if len(selection.FormatFilters) > 0 {
				filterParts = append(filterParts, fmt.Sprintf("formats %v", selection.FormatFilters))
			}

			if len(filterParts) > 0 {
				format.PrintFilter("Track filter", strings.Join(filterParts, ", "))
			}
		}
	} else if showFilterMessage {
		format.PrintInfo("No filter - muxing and extracting all subtitle tracks")
	}

	if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
		format.PrintError(fmt.Sprintf("File does not exist or is a directory: %s", inputFileName))
		return statErr
	}
	if !util.IsMKVFile(inputFileName) {
		format.PrintError(fmt.Sprintf("File is not an MKV file: %s", inputFileName))
		return errors.New("file is not an MKV file")
	}

	// Step 0: Get original track information to preserve track numbers
	originalMkvInfo, err := mkv.GetTrackInfo(inputFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error analyzing original file: %v", err))
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

	// Step 1: Create .mks file with only selected subtitle tracks
	mksFileName, mksErr := mkv.CreateSubtitlesMKS(inputFileName, selection, util.MatchesTrackSelection, outputConfig)
	if mksErr != nil {
		return mksErr
	}
	// Ensure cleanup of temporary .mks file
	defer mkv.CleanupTempFile(mksFileName)

	// Step 2: Get track information from the temporary .mks file
	mkvInfo, err := mkv.GetTrackInfo(mksFileName)
	if err != nil {
		format.PrintError(fmt.Sprintf("Error analyzing subtitle tracks: %v", err))
		return err
	}

	// Step 2: Extract subtitles
	format.PrintStep(2, "Extracting subtitle tracks...")

	var jobs []model.ExtractionJob
	mksTrackIndex := 0

	for _, track := range mkvInfo.Tracks {
		if track.Type == "subtitles" {
			// Use the corresponding original track based on order
			// The .mks file should contain tracks in the same order as they were selected
			var originalTrack model.MKVTrack
			if mksTrackIndex < len(selectedOriginalTracks) {
				originalTrack = selectedOriginalTracks[mksTrackIndex]
			} else {
				format.PrintWarning(fmt.Sprintf("Track index mismatch, using renumbered track info for track %d", track.Id))
				originalTrack = track
			}
			mksTrackIndex++

			outFileName := util.BuildSubtitlesFileNameWithConfig(inputFileName, originalTrack, outputConfig)

			jobs = append(jobs, model.ExtractionJob{
				Track:         track,
				OriginalTrack: originalTrack,
				OutFileName:   outFileName,
				MksFileName:   mksFileName,
			})
		}
	}

	// Execute optimized extraction using single mkvextract call per input file
	extractErr := mkv.ProcessTracks(jobs)
	if extractErr != nil {
		return extractErr
	}

	return nil
}

func main() {
	format.PrintTitle()

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
		inputFileName := strings.Join(args, " ")

		if ifs, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) || ifs.IsDir() {
			format.PrintError(fmt.Sprintf("File does not exist or is a directory: %s", inputFileName))
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}
		if !util.IsMKVFile(inputFileName) {
			format.PrintError(fmt.Sprintf("File is not an MKV file: %s", inputFileName))
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

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

	flags := struct {
		Extract        string `short:"x" long:"extract" description:"Extract subtitles from MKV file"`
		Info           string `short:"i" long:"info" description:"Display subtitle track information for MKV file"`
		Select         string `short:"s" long:"select" description:"Mixed selection of language codes and track IDs (e.g., 'eng,14,spa,16')"`
		OutputDir      string `short:"o" long:"output-dir" description:"Output directory for extracted subtitle files. If not specified, uses the same directory as the input file"`
		OutputTemplate string `short:"f" long:"format" description:"Custom filename template with placeholders: {basename}, {language}, {trackno}, {trackname}, {forced}, {default}, {extension}"`
	}{}

	_, cmdErr := gocmd.New(gocmd.Options{
		Name:        "subscalpelmkv",
		Description: "SubScalpelMKV - Extract subtitle tracks from MKV files quickly and precisely. Supports drag-and-drop: simply drag an MKV file onto the executable.",
		Version:     "1.0.0",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})

	if cmdErr != nil {
		format.PrintError(fmt.Sprintf("Error creating command: %v", cmdErr))
		return
	}

	if flags.Extract != "" && flags.Info != "" {
		format.PrintError("Cannot use both --extract and --info flags simultaneously")
		os.Exit(ErrCodeFailure)
	}

	if flags.Extract != "" {
		inputFileName := flags.Extract
		selectionFilter := cli.BuildSelectionFilter(flags.Select)

		outputConfig := model.OutputConfig{
			OutputDir: flags.OutputDir,
			Template:  flags.OutputTemplate,
			CreateDir: true, // Always create directory if it doesn't exist
		}

		if outputConfig.Template == "" {
			outputConfig.Template = model.DefaultOutputTemplate
		}

		err := processFile(inputFileName, selectionFilter, true, outputConfig)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else if flags.Info != "" {
		inputFileName := flags.Info
		err := cli.ShowFileInfo(inputFileName)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else {
		cli.ShowHelp()
		os.Exit(ErrCodeFailure)
	}

	os.Exit(ErrCodeSuccess)
}
