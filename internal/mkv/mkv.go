package mkv

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"subscalpelmkv/internal/model"
)

// SubtitleExtensionByCodec maps codec IDs to file extensions
var SubtitleExtensionByCodec = map[string]string{
	"S_TEXT/UTF8": "srt",
	"S_TEXT/ASS":  "ass",
	"S_HDMV/PGS":  "sup",
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
func ExtractSubtitles(inputFileName string, track model.MKVTrack, outFileName string) error {
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

// CleanupTempFile removes the temporary .mks file
func CleanupTempFile(fileName string) {
	if fileName != "" {
		if err := os.Remove(fileName); err != nil {
			// Silently ignore cleanup errors - not critical for user
		}
	}
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

// CreateSubtitlesMKS creates a .mks file containing only selected subtitle tracks from the input MKV file
func CreateSubtitlesMKS(inputFileName string, selection model.TrackSelection, matchesTrackSelection func(model.MKVTrack, model.TrackSelection) bool) (string, error) {
	// Create temporary .mks file path
	dir := path.Dir(inputFileName)
	baseName := strings.TrimSuffix(path.Base(inputFileName), path.Ext(inputFileName))
	mksFileName := path.Join(dir, baseName+".subtitles.mks")

	fmt.Println("Step 1: Creating temporary subtitle file...")

	// First, get track information from the original file to determine which tracks to include
	originalMkvInfo, err := GetTrackInfo(inputFileName)
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
