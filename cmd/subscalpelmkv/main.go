package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/devfacet/gocmd/v3"

	"subscalpelmkv/internal/batch"
	"subscalpelmkv/internal/cli"
	"subscalpelmkv/internal/config"
	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

const (
	ErrCodeSuccess = 0
	ErrCodeFailure = 1
)

var Version = "1.0.0"

// processFile handles the actual subtitle extraction logic
func processFile(inputFileName, languageFilter, exclusionFilter string, showFilterMessage bool, outputConfig model.OutputConfig, dryRun bool) error {
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

	// Parse exclusions if provided
	if exclusionFilter != "" {
		selection.Exclusions = cli.ParseTrackExclusion(exclusionFilter)
		if showFilterMessage {
			var exclusionParts []string
			if len(selection.Exclusions.LanguageCodes) > 0 {
				exclusionParts = append(exclusionParts, fmt.Sprintf("languages %v", selection.Exclusions.LanguageCodes))
			}
			if len(selection.Exclusions.TrackNumbers) > 0 {
				exclusionParts = append(exclusionParts, fmt.Sprintf("track IDs %v", selection.Exclusions.TrackNumbers))
			}
			if len(selection.Exclusions.FormatFilters) > 0 {
				exclusionParts = append(exclusionParts, fmt.Sprintf("formats %v", selection.Exclusions.FormatFilters))
			}

			if len(exclusionParts) > 0 {
				format.PrintFilter("Track exclusions", strings.Join(exclusionParts, ", "))
			}
		}
	}

	if _, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) {
		format.PrintError(fmt.Sprintf("File does not exist: %s", inputFileName))
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

	// For dry run mode, show what would be extracted without actually doing it
	if dryRun {
		if len(selectedOriginalTracks) == 0 {
			format.PrintWarning("No subtitle tracks match the selection criteria")
			return nil
		}

		format.PrintSubSection("Dry Run - Would Extract")
		format.PrintInfo(fmt.Sprintf("Would extract %d track(s) from: %s", len(selectedOriginalTracks), filepath.Base(inputFileName)))

		for _, track := range selectedOriginalTracks {
			outFileName := util.BuildSubtitlesFileNameWithConfig(inputFileName, track, outputConfig)

			// Get codec type for display
			codecType := "Unknown"
			if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
				codecType = strings.ToUpper(ext)
			}

			// Build track details string
			trackDetails := fmt.Sprintf("Track %d (%s)", track.Properties.Number, track.Properties.Language)
			if track.Properties.TrackName != "" {
				trackDetails += fmt.Sprintf(" - %s", track.Properties.TrackName)
			}

			// Add format and attributes
			attributes := []string{codecType}
			if track.Properties.Forced {
				attributes = append(attributes, "forced")
			}
			if track.Properties.Default {
				attributes = append(attributes, "default")
			}

			format.BorderColor.Print("  ")
			format.BaseHighlight.Print("▪")
			fmt.Print(" ")
			format.BaseFg.Println(fmt.Sprintf("%s [%s]", trackDetails, strings.Join(attributes, ", ")))
			format.PrintExample(fmt.Sprintf("    → %s", outFileName))
		}

		return nil
	}

	fmt.Println()
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

	fmt.Println()
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

// processBatch handles batch processing of multiple MKV files
func processBatch(pattern, languageFilter, exclusionFilter string, showFilterMessage bool, outputConfig model.OutputConfig, dryRun bool) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		format.PrintError(fmt.Sprintf("Invalid glob pattern: %v", err))
		return err
	}

	if len(files) == 0 {
		format.PrintError(fmt.Sprintf("No files found matching pattern: %s", pattern))
		return errors.New("no files found")
	}

	// Filter to only MKV files
	mkvFiles, err := util.ValidateAndFilterMKVFiles(files)
	if err != nil {
		format.PrintError(fmt.Sprintf("No MKV files found matching pattern: %s", pattern))
		return err
	}

	format.PrintInfo(fmt.Sprintf("Found %d MKV file(s) to process", len(mkvFiles)))

	if showFilterMessage && languageFilter != "" {
		selection := cli.ParseTrackSelection(languageFilter)
		exclusion := cli.ParseTrackExclusion(exclusionFilter)
		selectionResult := cli.ProcessSelectionForBatch(selection, exclusion)
		if selectionResult.Message != "" {
			format.PrintFilter("Batch filter", selectionResult.Message)
		}
	} else if showFilterMessage {
		format.PrintInfo("No filter - extracting all subtitle tracks from each file")
	}

	// Use the new batch processor
	processor := batch.NewProcessor(mkvFiles, outputConfig, dryRun)
	result, err := processor.Process(processFile, languageFilter, exclusionFilter)
	if err != nil {
		return err
	}

	processor.PrintSummary(result)

	if result.ErrorCount > 0 {
		return fmt.Errorf("batch processing completed with %d errors", result.ErrorCount)
	}

	return nil
}

// handleBatchDragAndDrop handles drag-and-drop of multiple MKV files
func handleBatchDragAndDrop(mkvFiles []string, outputConfig model.OutputConfig) error {
	format.PrintInfo(fmt.Sprintf("Batch drag-and-drop detected: %d MKV files", len(mkvFiles)))

	// Analyze each file to gather subtitle information
	batchFileInfos := batch.AnalyzeFiles(mkvFiles)

	// Display all files using the same visual style as subtitle tracks
	cli.DisplayBatchFiles(batchFileInfos)

	// Ask user if they want to extract all tracks or make a selection
	extractAll := cli.AskUserConfirmation()

	// Collect all available track numbers from all files for validation
	var allAvailableTracks []int
	trackSet := make(map[int]bool)
	for _, fileInfo := range batchFileInfos {
		if !fileInfo.HasError {
			// Get track info for this file
			mkvInfo, err := mkv.GetTrackInfo(fileInfo.FilePath)
			if err == nil {
				for _, track := range mkvInfo.Tracks {
					if track.Type == "subtitles" {
						if !trackSet[track.Properties.Number] {
							trackSet[track.Properties.Number] = true
							allAvailableTracks = append(allAvailableTracks, track.Properties.Number)
						}
					}
				}
			}
		}
	}

	// Process selection and exclusion using the shared function
	selectionResult, err := cli.ProcessSelectionAndExclusion(extractAll, allAvailableTracks)
	if err != nil {
		fmt.Println("Press enter to exit...")
		fmt.Scanln()
		return nil
	}

	if selectionResult.Message != "" {
		format.PrintSubSection(selectionResult.Title)
		format.PrintInfo(selectionResult.Message)
	}

	// Filter out files that had analysis errors and prepare valid files for processing
	validFiles := batch.FilterValidFiles(batchFileInfos)

	if len(validFiles) == 0 {
		format.PrintError("No valid MKV files to process")
		fmt.Println("Press enter to exit...")
		fmt.Scanln()
		return fmt.Errorf("no valid files to process")
	}

	// Use the batch processor for consistent handling
	processor := batch.NewProcessor(validFiles, outputConfig, false)
	result, _ := processor.Process(processFile, selectionResult.LanguageFilter, selectionResult.ExclusionFilter)
	processor.PrintSummary(result)

	fmt.Println("Press enter to exit...")
	fmt.Scanln()

	if result.ErrorCount > 0 {
		return fmt.Errorf("batch processing completed with %d errors", result.ErrorCount)
	}

	return nil
}

func main() {
	format.PrintTitleWithVersion(Version)

	args := os.Args[1:]

	// Check for help flags first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			cli.ShowHelp()
			os.Exit(ErrCodeSuccess)
		}
	}

	// Check if -o flag is used without arguments and handle it specially
	hasOutputFlagWithoutValue := false
	modifiedArgs := make([]string, len(args))
	copy(modifiedArgs, args)

	for i, arg := range args {
		if arg == "-o" || arg == "--output-dir" {
			// Check if next argument exists and doesn't start with '-'
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				hasOutputFlagWithoutValue = true
				// Insert a special marker value that gocmd can parse
				modifiedArgs = append(modifiedArgs[:i+1], append([]string{"__BASENAME_SUBTITLES__"}, modifiedArgs[i+1:]...)...)
				break
			}
		}
	}

	// Replace the original os.Args with our modified version for gocmd
	originalArgs := os.Args
	os.Args = append([]string{os.Args[0]}, modifiedArgs...)

	// Detect execution mode: drag-and-drop vs CLI
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Use the new discovery function
		validMKVFiles, err := util.DiscoverMKVFiles(args)
		if err != nil {
			format.PrintError(fmt.Sprintf("Error discovering MKV files: %v", err))
			fmt.Println("Press enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		// If we found multiple valid MKV files (from files or directories), handle as batch
		if len(validMKVFiles) > 1 {
			defaultOutputConfig := util.BuildOutputConfig("", "", false, false)
			err = handleBatchDragAndDrop(validMKVFiles, defaultOutputConfig)
			if err != nil {
				os.Exit(ErrCodeFailure)
			}
			os.Exit(ErrCodeSuccess)
		}

		// If we found exactly one valid file, process it
		if len(validMKVFiles) == 1 {
			defaultOutputConfig := util.BuildOutputConfig("", "", false, false)
			err = cli.HandleDragAndDropModeWithConfig(validMKVFiles[0], processFile, defaultOutputConfig)
			if err != nil {
				os.Exit(ErrCodeFailure)
			}
			os.Exit(ErrCodeSuccess)
		}

		// If no valid files found, try the traditional approach (joining with spaces for filenames with spaces)
		inputFileName := strings.Join(args, " ")

		if _, statErr := os.Stat(inputFileName); os.IsNotExist(statErr) {
			format.PrintError(fmt.Sprintf("File does not exist: %s", inputFileName))
			fmt.Println("Press enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		// Check if it's a directory
		if info, _ := os.Stat(inputFileName); info.IsDir() {
			format.PrintInfo(fmt.Sprintf("Scanning directory: %s", inputFileName))
			files, err := util.FindMKVFilesInDirectory(inputFileName)
			if err != nil {
				format.PrintError(fmt.Sprintf("Error scanning directory: %v", err))
				fmt.Println("Press enter to exit...")
				fmt.Scanln()
				os.Exit(ErrCodeFailure)
			}

			if len(files) == 0 {
				format.PrintError("No MKV files found in the directory")
				fmt.Println("Press enter to exit...")
				fmt.Scanln()
				os.Exit(ErrCodeFailure)
			}

			defaultOutputConfig := util.BuildOutputConfig("", "", false, false)

			if len(files) == 1 {
				err = cli.HandleDragAndDropModeWithConfig(files[0], processFile, defaultOutputConfig)
				if err != nil {
					os.Exit(ErrCodeFailure)
				}
			} else {
				err = handleBatchDragAndDrop(files, defaultOutputConfig)
				if err != nil {
					os.Exit(ErrCodeFailure)
				}
			}
			os.Exit(ErrCodeSuccess)
		}

		if !util.IsMKVFile(inputFileName) {
			format.PrintError(fmt.Sprintf("File is not an MKV file: %s", inputFileName))
			fmt.Println("Press enter to exit...")
			fmt.Scanln()
			os.Exit(ErrCodeFailure)
		}

		defaultOutputConfig := util.BuildOutputConfig("", "", false, false)
		err = cli.HandleDragAndDropModeWithConfig(inputFileName, processFile, defaultOutputConfig)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
		os.Exit(ErrCodeSuccess)
	}

	flags := struct {
		Extract        string `short:"x" long:"extract" description:"Extract subtitles from MKV file"`
		Batch          string `short:"b" long:"batch" description:"Extract subtitles from multiple MKV files using glob pattern (e.g., '*.mkv', 'Season 1/*.mkv')"`
		Info           string `short:"i" long:"info" description:"Display subtitle track information for MKV file"`
		Select         string `short:"s" long:"select" description:"Mixed selection of language codes and track IDs (e.g., 'eng,14,spa,16')"`
		Exclude        string `short:"e" long:"exclude" description:"Mixed exclusion of language codes, track IDs, and formats (e.g., 'chi,15,sup')"`
		OutputDir      string `short:"o" long:"output-dir" description:"Output directory for extracted subtitle files. If not specified, uses the same directory as the input file"`
		OutputTemplate string `short:"f" long:"format" description:"Custom filename template with placeholders: {basename}, {language}, {trackno}, {trackname}, {forced}, {default}, {extension}"`
		DryRun         bool   `short:"d" long:"dry-run" description:"Show what would be extracted without performing extraction"`
		UseConfig      bool   `short:"c" long:"config" description:"Use default configuration profile"`
		Profile        string `short:"p" long:"profile" description:"Use named configuration profile"`
	}{}

	_, cmdErr := gocmd.New(gocmd.Options{
		Name:        "subscalpelmkv",
		Description: "SubScalpelMKV - Extract subtitle tracks from MKV files. Use CLI or drag-and-drop directories and MKV files",
		Version:     "1.0.0",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})

	if cmdErr != nil {
		format.PrintError(fmt.Sprintf("Error creating command: %v", cmdErr))
		return
	}

	// Load configuration if requested
	var appliedConfig *config.AppliedConfig
	if flags.UseConfig || flags.Profile != "" {
		cfg, err := config.LoadConfigWithFallback()
		if err != nil {
			format.PrintError(fmt.Sprintf("Error loading configuration: %v", err))
			os.Exit(ErrCodeFailure)
		}

		if flags.Profile != "" {
			appliedConfig, err = cfg.ApplyProfile(flags.Profile)
			if err != nil {
				format.PrintError(fmt.Sprintf("Error applying profile '%s': %v", flags.Profile, err))
				os.Exit(ErrCodeFailure)
			}
		} else {
			appliedConfig = cfg.ApplyDefaults()
		}

		// Merge configuration with CLI flags (CLI flags take precedence)
		cliFlags := config.CLIFlags{
			OutputTemplate: flags.OutputTemplate,
			OutputDir:      flags.OutputDir,
		}

		// Parse languages from Select flag if provided
		if flags.Select != "" {
			selection := cli.ParseTrackSelection(flags.Select)
			cliFlags.Languages = selection.LanguageCodes
		}

		// Parse exclusions from Exclude flag if provided
		if flags.Exclude != "" {
			exclusion := cli.ParseTrackExclusion(flags.Exclude)
			var exclusionParts []string
			exclusionParts = append(exclusionParts, exclusion.LanguageCodes...)
			for _, trackNum := range exclusion.TrackNumbers {
				exclusionParts = append(exclusionParts, strconv.Itoa(trackNum))
			}
			exclusionParts = append(exclusionParts, exclusion.FormatFilters...)
			cliFlags.Exclusions = exclusionParts
		}

		appliedConfig = appliedConfig.MergeWithCLI(cliFlags)

		// Apply config values back to flags if they weren't set via CLI
		if flags.OutputTemplate == "" && appliedConfig.OutputTemplate != "" {
			flags.OutputTemplate = appliedConfig.OutputTemplate
		}
		if flags.OutputDir == "" && appliedConfig.OutputDir != "" {
			flags.OutputDir = appliedConfig.OutputDir
		}
		if flags.Select == "" && len(appliedConfig.Languages) > 0 {
			flags.Select = strings.Join(appliedConfig.Languages, ",")
		}
		if flags.Exclude == "" && len(appliedConfig.Exclusions) > 0 {
			flags.Exclude = strings.Join(appliedConfig.Exclusions, ",")
		}
	}

	if (flags.Extract != "" && flags.Info != "") ||
		(flags.Extract != "" && flags.Batch != "") ||
		(flags.Info != "" && flags.Batch != "") {
		format.PrintError("Cannot use multiple processing flags simultaneously (--extract, --batch, --info)")
		os.Exit(ErrCodeFailure)
	}

	if flags.Extract != "" {
		inputFileName := flags.Extract
		selectionFilter := cli.BuildSelectionFilter(flags.Select)

		outputConfig := util.BuildOutputConfig(flags.OutputDir, flags.OutputTemplate, hasOutputFlagWithoutValue, false)

		// Resolve special output directory for single file
		if outputConfig.OutputDir == "__BASENAME_SUBTITLES__" {
			outputConfig.OutputDir = util.ResolveOutputDirectory(outputConfig.OutputDir, inputFileName)
		}

		err := processFile(inputFileName, selectionFilter, flags.Exclude, true, outputConfig, flags.DryRun)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else if flags.Batch != "" {
		pattern := flags.Batch
		selectionFilter := cli.BuildSelectionFilter(flags.Select)

		outputConfig := util.BuildOutputConfig(flags.OutputDir, flags.OutputTemplate, hasOutputFlagWithoutValue, true)

		err := processBatch(pattern, selectionFilter, flags.Exclude, true, outputConfig, flags.DryRun)
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

	// Restore original args
	os.Args = originalArgs
}
