package mkv

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

// printExtractedTrackSuccess prints the extraction success message in a two-line format matching dry-run style
func printExtractedTrackSuccess(trackNumber int, track model.MKVTrack, outFileName string) {
	// Get codec type for display
	codecType := "Unknown"
	if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
		codecType = strings.ToUpper(ext)
	}

	// Build track details string
	trackDetails := fmt.Sprintf("Track %d (%s)", trackNumber, track.Properties.Language)
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

	// First line: Track details with checkmark
	format.SuccessColor.Print("  ✓ ")
	format.BaseFg.Println(fmt.Sprintf("%s [%s]", trackDetails, strings.Join(attributes, ", ")))

	// Second line: Output path with arrow
	format.PrintExample(fmt.Sprintf("    → %s", outFileName))
	fmt.Println()
}

// GetTrackInfo gets track information from an MKV file using mkvmerge -J
func GetTrackInfo(inputFileName string) (*model.MKVInfo, error) {
	out, cmdErr := exec.Command("mkvmerge", "-J", inputFileName).Output()
	if cmdErr != nil {
		return nil, fmt.Errorf("error analyzing tracks: %v", cmdErr)
	}

	var mkvInfo model.MKVInfo
	jsonErr := json.Unmarshal(out, &mkvInfo)
	if jsonErr != nil {
		return nil, fmt.Errorf("error parsing track information: %v", jsonErr)
	}

	if !(strings.ToLower(strings.TrimSpace(mkvInfo.Container.Type)) == "matroska") {
		return nil, errors.New("file is not a valid Matroska container")
	}

	return &mkvInfo, nil
}

// ExtractSubtitles extracts a subtitle track from an MKV file
func ExtractSubtitles(inputFileName string, track model.MKVTrack, outFileName string, originalTrackNumber int) error {
	cmd := exec.Command(
		"mkvextract",
		fmt.Sprintf("%v", inputFileName),
		"tracks",
		fmt.Sprintf("%d:%v", track.Id, outFileName),
	)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		format.PrintError(fmt.Sprintf("Error extracting track %d: %v", track.Id, cmdErr))
		fmt.Println(string(output))
		return cmdErr
	}

	// Handle special case for S_VOBSUB which creates both .idx and .sub files
	if track.Properties.CodecId == "S_VOBSUB" {
		// For VOBSUB, mkvextract creates both .idx and .sub files automatically
		// The output filename should have .sub extension, and .idx will be created alongside it
		baseFileName := strings.TrimSuffix(outFileName, filepath.Ext(outFileName))
		idxFileName := baseFileName + ".idx"
		subFileName := baseFileName + ".sub"
		// For VOBSUB, show both files in the output path
		combinedOutput := fmt.Sprintf("%s + %s", filepath.Base(idxFileName), filepath.Base(subFileName))
		printExtractedTrackSuccess(originalTrackNumber, track, combinedOutput)
	} else {
		printExtractedTrackSuccess(originalTrackNumber, track, outFileName)
	}
	return nil
}

// TrackExtractionInfo represents information needed to extract a single track
type TrackExtractionInfo struct {
	Track         model.MKVTrack
	OriginalTrack model.MKVTrack
	OutFileName   string
}

// ExtractMultipleSubtitles extracts multiple subtitle tracks from a single input file in one mkvextract call
func ExtractMultipleSubtitles(inputFileName string, tracks []TrackExtractionInfo) error {
	if len(tracks) == 0 {
		return nil
	}

	args := []string{inputFileName, "tracks"}

	for _, trackInfo := range tracks {
		trackPair := fmt.Sprintf("%d:%s", trackInfo.Track.Id, trackInfo.OutFileName)
		args = append(args, trackPair)
	}

	cmd := exec.Command("mkvextract", args...)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		format.PrintError(fmt.Sprintf("Error extracting tracks: %v", cmdErr))
		fmt.Println(string(output))
		return cmdErr
	}

	for _, trackInfo := range tracks {
		track := trackInfo.Track
		originalTrack := trackInfo.OriginalTrack
		outFileName := trackInfo.OutFileName

		// Handle special case for S_VOBSUB which creates both .idx and .sub files
		if track.Properties.CodecId == "S_VOBSUB" {
			// For VOBSUB, mkvextract creates both .idx and .sub files automatically
			// The output filename should have .sub extension, and .idx will be created alongside it
			baseFileName := strings.TrimSuffix(outFileName, filepath.Ext(outFileName))
			idxFileName := baseFileName + ".idx"
			subFileName := baseFileName + ".sub"
			// For VOBSUB, show both files in the output path
			combinedOutput := fmt.Sprintf("%s + %s", filepath.Base(idxFileName), filepath.Base(subFileName))
			printExtractedTrackSuccess(originalTrack.Properties.Number, track, combinedOutput)
		} else {
			printExtractedTrackSuccess(originalTrack.Properties.Number, track, outFileName)
		}
	}

	return nil
}

// CleanupTempFile removes the temporary .mks file
func CleanupTempFile(fileName string) {
	if fileName != "" {
		if err := os.Remove(fileName); err != nil {
			// Silently ignore cleanup errors - not critical for user
		}
	}
}

// CreateSubtitlesMKS creates a .mks file containing only selected subtitle tracks from the input MKV file
func CreateSubtitlesMKS(inputFileName string, selection model.TrackSelection, matchesTrackSelection func(model.MKVTrack, model.TrackSelection) bool, outputConfig model.OutputConfig) (string, error) {
	// Create temporary .mks file path - use the same directory as the output files
	var dir string
	if outputConfig.OutputDir != "" {
		dir = outputConfig.OutputDir
		// Always create output directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			format.PrintWarning(fmt.Sprintf("Could not create output directory %s: %v", dir, err))
			// Fall back to input file directory
			dir = filepath.Dir(inputFileName)
		}
	} else {
		dir = filepath.Dir(inputFileName)
	}
	baseName := strings.TrimSuffix(filepath.Base(inputFileName), filepath.Ext(inputFileName))
	mksFileName := filepath.Join(dir, baseName+".subtitles.mks")

	format.PrintStep(1, "Preparing selected tracks for extraction...")

	// First, get track information from the original file to determine which tracks to include
	originalMkvInfo, err := GetTrackInfo(inputFileName)
	if err != nil {
		return "", fmt.Errorf("failed to analyze original file: %v", err)
	}

	// Build list of subtitle track IDs that match the selection criteria
	var selectedTrackIDs []string
	for _, track := range originalMkvInfo.Tracks {
		if track.Type == "subtitles" {
			if matchesTrackSelection(track, selection) {
				selectedTrackIDs = append(selectedTrackIDs, strconv.Itoa(track.Id))
			}
		}
	}

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

	// Add subtitle track selection - always specify which tracks to include when we have selections or exclusions
	hasSelectionCriteria := len(selection.LanguageCodes) > 0 || len(selection.TrackNumbers) > 0 || len(selection.FormatFilters) > 0
	hasExclusionCriteria := len(selection.Exclusions.LanguageCodes) > 0 || len(selection.Exclusions.TrackNumbers) > 0 || len(selection.Exclusions.FormatFilters) > 0
	
	if hasSelectionCriteria || hasExclusionCriteria {
		subtitleTracks := strings.Join(selectedTrackIDs, ",")
		args = append(args, "--subtitle-tracks", subtitleTracks)

		// Build display list using track.Properties.Number for user-friendly output
		var displayTrackNumbers []string
		for _, track := range originalMkvInfo.Tracks {
			if track.Type == "subtitles" {
				if matchesTrackSelection(track, selection) {
					displayTrackNumbers = append(displayTrackNumbers, strconv.Itoa(track.Properties.Number))
				}
			}
		}
		format.PrintInfo(fmt.Sprintf("Including subtitle tracks: %s", strings.Join(displayTrackNumbers, ",")))
	}

	args = append(args, inputFileName)
	cmd := exec.Command("mkvmerge", args...)

	// Set up pipe to capture stdout for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	// Also capture stderr to prevent blocking if mkvmerge writes errors/warnings
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mkvmerge: %v", err)
	}

	// Start a goroutine to consume stderr to prevent blocking
	var stderrOutput strings.Builder
	go func() {
		scanner := bufio.NewScanner(stderr)
		// Increase buffer size for stderr as well
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		
		for scanner.Scan() {
			stderrOutput.WriteString(scanner.Text() + "\n")
		}
	}()

	// Hide cursor for cleaner progress display
	fmt.Print("\033[?25l")

	// Show initial 0% progress bar immediately
	util.ShowProgressBar(0)

	// Create a ticker to update elapsed time every 100ms
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)
	
	// Start goroutine to update elapsed time
	go func() {
		for {
			select {
			case <-ticker.C:
				util.UpdateElapsedTime()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Monitor stdout for progress information
	scanner := bufio.NewScanner(stdout)
	// Increase buffer size to handle potentially long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Allow up to 1MB lines
	
	for scanner.Scan() {
		line := scanner.Text()

		if percentage, isProgress := util.ParseProgressLine(line); isProgress {
			util.ShowProgressBar(percentage)
		}
	}

	// Stop the ticker
	done <- true
	cmdErr := cmd.Wait()

	// Show cursor again
	fmt.Print("\033[?25h")

	if cmdErr != nil {
		// Clear the progress line before showing error
		fmt.Print("\r\033[K")
		format.PrintError(fmt.Sprintf("Error creating temporary subtitle file: %v", cmdErr))
		// If there was stderr output, display it for debugging
		if stderrStr := stderrOutput.String(); stderrStr != "" {
			format.PrintError(fmt.Sprintf("mkvmerge stderr: %s", strings.TrimSpace(stderrStr)))
		}
		return "", cmdErr
	}

	return mksFileName, nil
}

// ProcessTracks groups extraction jobs by input file and processes them efficiently
func ProcessTracks(jobs []model.ExtractionJob) error {
	if len(jobs) == 0 {
		format.PrintWarning("No subtitle tracks to extract")
		return nil
	}

	// Group jobs by input file (MksFileName in this case, since that's the actual input for extraction)
	jobsByInputFile := make(map[string][]TrackExtractionInfo)

	for _, job := range jobs {
		inputFile := job.MksFileName
		trackInfo := TrackExtractionInfo{
			Track:         job.Track,
			OriginalTrack: job.OriginalTrack,
			OutFileName:   job.OutFileName,
		}
		jobsByInputFile[inputFile] = append(jobsByInputFile[inputFile], trackInfo)
	}

	// Process each input file with a single mkvextract call
	successCount := 0

	for inputFile, tracks := range jobsByInputFile {
		err := ExtractMultipleSubtitles(inputFile, tracks)
		if err != nil {
			format.PrintError(fmt.Sprintf("Error extracting tracks from %s: %v", inputFile, err))
			return err
		}
		successCount += len(tracks)
	}

	if successCount == 0 {
		format.PrintWarning("No subtitle tracks were extracted")
	} else {
		format.PrintSuccess(fmt.Sprintf("Successfully extracted %d subtitle track(s)", successCount))
	}

	return nil
}
