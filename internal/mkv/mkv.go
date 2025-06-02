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

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

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
		format.PrintSuccess(fmt.Sprintf("Extracted track ID %d (%s) -> %s + %s", originalTrackNumber, track.Properties.Language,
			filepath.Base(idxFileName), filepath.Base(subFileName)))
	} else {
		format.PrintSuccess(fmt.Sprintf("Extracted track ID %d (%s) -> %s", originalTrackNumber, track.Properties.Language, outFileName))
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
			format.PrintSuccess(fmt.Sprintf("Extracted track ID %d (%s) -> %s + %s", originalTrack.Properties.Number, track.Properties.Language,
				filepath.Base(idxFileName), filepath.Base(subFileName)))
		} else {
			format.PrintSuccess(fmt.Sprintf("Extracted track ID %d (%s) -> %s", originalTrack.Properties.Number, track.Properties.Language, outFileName))
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

	// Add subtitle track selection - only include matching tracks
	if len(selection.LanguageCodes) > 0 || len(selection.TrackNumbers) > 0 || len(selection.FormatFilters) > 0 {
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

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mkvmerge: %v", err)
	}

	// Hide cursor for cleaner progress display
	fmt.Print("\033[?25l")

	// Monitor stdout for progress information
	scanner := bufio.NewScanner(stdout)

	// Show initial 0% progress bar immediately
	util.ShowProgressBar(0)

	for scanner.Scan() {
		line := scanner.Text()

		if percentage, isProgress := util.ParseProgressLine(line); isProgress {
			util.ShowProgressBar(percentage)
		}
	}

	cmdErr := cmd.Wait()

	// Show cursor again
	fmt.Print("\033[?25h")

	if cmdErr != nil {
		// Clear the progress line before showing error
		fmt.Print("\r\033[K")
		format.PrintError(fmt.Sprintf("Error creating temporary subtitle file: %v", cmdErr))
		return "", cmdErr
	}

	fmt.Println()
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

	fmt.Println()
	if successCount == 0 {
		format.PrintWarning("No subtitle tracks were extracted")
	} else {
		format.PrintSuccess(fmt.Sprintf("Successfully extracted %d subtitle track(s)", successCount))
	}

	return nil
}
