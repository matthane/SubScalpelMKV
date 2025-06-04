package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// processBatch handles batch processing of multiple MKV files
func processBatch(pattern, languageFilter string, showFilterMessage bool, outputConfig model.OutputConfig) error {
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
	var mkvFiles []string
	for _, file := range files {
		if info, err := os.Stat(file); err == nil && !info.IsDir() && util.IsMKVFile(file) {
			mkvFiles = append(mkvFiles, file)
		}
	}

	if len(mkvFiles) == 0 {
		format.PrintError(fmt.Sprintf("No MKV files found matching pattern: %s", pattern))
		return errors.New("no MKV files found")
	}

	format.PrintInfo(fmt.Sprintf("Found %d MKV file(s) to process", len(mkvFiles)))
	
	if showFilterMessage && languageFilter != "" {
		var selection model.TrackSelection = cli.ParseTrackSelection(languageFilter)
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
			format.PrintFilter("Batch filter", strings.Join(filterParts, ", "))
		}
	} else if showFilterMessage {
		format.PrintInfo("No filter - extracting all subtitle tracks from each file")
	}

	// Process each file
	successCount := 0
	errorCount := 0
	
	for i, file := range mkvFiles {
		format.PrintSubSection(fmt.Sprintf("Processing file %d/%d: %s", i+1, len(mkvFiles), filepath.Base(file)))
		
		err := processFile(file, languageFilter, false, outputConfig)
		if err != nil {
			format.PrintError(fmt.Sprintf("Failed to process %s: %v", file, err))
			errorCount++
		} else {
			format.PrintSuccess(fmt.Sprintf("Successfully processed %s", filepath.Base(file)))
			successCount++
		}
		
		// Add spacing between files except for the last one
		if i < len(mkvFiles)-1 {
			fmt.Println()
		}
	}

	// Print summary
	fmt.Println()
	format.PrintSubSection("Batch Processing Summary")
	format.PrintInfo(fmt.Sprintf("Total files: %d", len(mkvFiles)))
	format.PrintSuccess(fmt.Sprintf("Successfully processed: %d", successCount))
	if errorCount > 0 {
		format.PrintError(fmt.Sprintf("Failed to process: %d", errorCount))
	}

	if errorCount > 0 {
		return fmt.Errorf("batch processing completed with %d errors", errorCount)
	}

	return nil
}

// handleBatchDragAndDrop handles drag-and-drop of multiple MKV files
func handleBatchDragAndDrop(mkvFiles []string, outputConfig model.OutputConfig) error {
	format.PrintInfo(fmt.Sprintf("Batch drag-and-drop detected: %d MKV files", len(mkvFiles)))
	
	// Analyze each file to gather subtitle information
	var batchFileInfos []model.BatchFileInfo
	for _, file := range mkvFiles {
		fileInfo := model.BatchFileInfo{
			FileName: filepath.Base(file),
			FilePath: file,
		}
		
		// Try to get track information for this file
		mkvInfo, err := mkv.GetTrackInfo(file)
		if err != nil {
			fileInfo.HasError = true
			fileInfo.ErrorMessage = fmt.Sprintf("Failed to analyze: %v", err)
		} else {
			// Count subtitle tracks and gather language codes and formats
			languageSet := make(map[string]bool)
			formatSet := make(map[string]bool)
			
			for _, track := range mkvInfo.Tracks {
				if track.Type == "subtitles" {
					fileInfo.SubtitleCount++
					
					// Collect language codes
					if track.Properties.Language != "" {
						languageSet[track.Properties.Language] = true
					}
					
					// Collect formats
					if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
						formatSet[ext] = true
					}
				}
			}
			
			// Convert sets to slices
			for lang := range languageSet {
				fileInfo.LanguageCodes = append(fileInfo.LanguageCodes, lang)
			}
			for format := range formatSet {
				fileInfo.SubtitleFormats = append(fileInfo.SubtitleFormats, format)
			}
		}
		
		batchFileInfos = append(batchFileInfos, fileInfo)
	}
	
	// Display all files using the same visual style as subtitle tracks
	cli.DisplayBatchFiles(batchFileInfos)
	
	// Ask user if they want to extract all tracks or make a selection
	extractAll := cli.AskUserConfirmation()
	
	var languageFilter string
	if !extractAll {
		selectionInput := cli.AskTrackSelection()
		selection := cli.ParseTrackSelection(selectionInput)
		
		if len(selection.LanguageCodes) == 0 && len(selection.TrackNumbers) == 0 && len(selection.FormatFilters) == 0 {
			format.PrintWarning("No valid language codes, track IDs, or format filters provided. Exiting.")
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			return nil
		}
		
		// Convert to comma-separated string for processFile function
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
		format.PrintInfo("Extracting all subtitle tracks from each file...")
	}
	fmt.Println()
	
	// Filter out files that had analysis errors and prepare valid files for processing
	var validFiles []string
	for _, fileInfo := range batchFileInfos {
		if !fileInfo.HasError {
			validFiles = append(validFiles, fileInfo.FilePath)
		}
	}
	
	if len(validFiles) == 0 {
		format.PrintError("No valid MKV files to process")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return fmt.Errorf("no valid files to process")
	}
	
	// Process each valid file
	successCount := 0
	errorCount := 0
	
	for i, file := range validFiles {
		format.PrintSubSection(fmt.Sprintf("Processing file %d/%d: %s", i+1, len(validFiles), filepath.Base(file)))
		
		err := processFile(file, languageFilter, false, outputConfig)
		if err != nil {
			format.PrintError(fmt.Sprintf("Failed to process %s: %v", file, err))
			errorCount++
		} else {
			format.PrintSuccess(fmt.Sprintf("Successfully processed %s", filepath.Base(file)))
			successCount++
		}
		
		// Add spacing between files except for the last one
		if i < len(validFiles)-1 {
			fmt.Println()
		}
	}
	
	// Print summary
	fmt.Println()
	format.PrintSubSection("Batch Processing Summary")
	format.PrintInfo(fmt.Sprintf("Total files: %d", len(validFiles)))
	format.PrintSuccess(fmt.Sprintf("Successfully processed: %d", successCount))
	if errorCount > 0 {
		format.PrintError(fmt.Sprintf("Failed to process: %d", errorCount))
	}
	
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	
	if errorCount > 0 {
		return fmt.Errorf("batch processing completed with %d errors", errorCount)
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
		// Check if we have multiple separate files vs one file with spaces
		var validMKVFiles []string
		
		// First, try each argument as a separate file
		for _, arg := range args {
			if info, err := os.Stat(arg); err == nil && !info.IsDir() && util.IsMKVFile(arg) {
				validMKVFiles = append(validMKVFiles, arg)
			}
		}
		
		// If we found multiple valid MKV files, handle as batch drag-and-drop
		if len(validMKVFiles) > 1 {
			defaultOutputConfig := model.OutputConfig{
				OutputDir: "",
				Template:  model.DefaultOutputTemplate,
				CreateDir: false,
			}
			err := handleBatchDragAndDrop(validMKVFiles, defaultOutputConfig)
			if err != nil {
				os.Exit(ErrCodeFailure)
			}
			os.Exit(ErrCodeSuccess)
		}
		
		// If we found exactly one valid file, or no valid separate files,
		// try the traditional approach (joining with spaces for filenames with spaces)
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
		Batch          string `short:"b" long:"batch" description:"Extract subtitles from multiple MKV files using glob pattern (e.g., '*.mkv', 'Season 1/*.mkv')"`
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

	if (flags.Extract != "" && flags.Info != "") || 
	   (flags.Extract != "" && flags.Batch != "") || 
	   (flags.Info != "" && flags.Batch != "") {
		format.PrintError("Cannot use multiple processing flags simultaneously (--extract, --batch, --info)")
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

		// Handle special case where -o is used without arguments
		if hasOutputFlagWithoutValue || flags.OutputDir == "__BASENAME_SUBTITLES__" {
			baseName := strings.TrimSuffix(filepath.Base(inputFileName), filepath.Ext(inputFileName))
			outputConfig.OutputDir = filepath.Join(filepath.Dir(inputFileName), baseName+"-subtitles")
		}

		if outputConfig.Template == "" {
			outputConfig.Template = model.DefaultOutputTemplate
		}

		err := processFile(inputFileName, selectionFilter, true, outputConfig)
		if err != nil {
			os.Exit(ErrCodeFailure)
		}
	} else if flags.Batch != "" {
		pattern := flags.Batch
		selectionFilter := cli.BuildSelectionFilter(flags.Select)

		outputConfig := model.OutputConfig{
			OutputDir: flags.OutputDir,
			Template:  flags.OutputTemplate,
			CreateDir: true, // Always create directory if it doesn't exist
		}

		// Handle special case where -o is used without arguments
		// For batch mode, we'll create individual {basename}-subtitles directories for each file
		if hasOutputFlagWithoutValue || flags.OutputDir == "__BASENAME_SUBTITLES__" {
			outputConfig.OutputDir = "BATCH_BASENAME_SUBTITLES" // Special marker for batch mode
		}

		if outputConfig.Template == "" {
			outputConfig.Template = model.DefaultOutputTemplate
		}

		err := processBatch(pattern, selectionFilter, true, outputConfig)
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
