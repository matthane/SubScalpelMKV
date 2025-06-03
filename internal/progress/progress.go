package progress

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"subscalpelmkv/internal/format"
)

var (
	lastPercent int
	startTime   time.Time
	once        sync.Once
	barWidth    = 60
)

// ProgressTheme defines the characters used for the progress bar
type ProgressTheme struct {
	Saucer        string
	SaucerHead    string
	SaucerPadding string
	BarStart      string
	BarEnd        string
}

var defaultTheme = ProgressTheme{
	Saucer:        "█",
	SaucerHead:    "█",
	SaucerPadding: "░",
	BarStart:      "▐",
	BarEnd:        "▌",
}

// ShowProgressBar displays a progress bar based on percentage
func ShowProgressBar(percentage int) {
	// Initialize only once
	once.Do(func() {
		startTime = time.Now()
		lastPercent = 0
		// Don't print "Muxing subtitle tracks" here - let the caller handle the initial message
	})

	renderProgressBar(percentage)
	lastPercent = percentage

	if percentage >= 100 {
		elapsed := time.Since(startTime)
		fmt.Printf("\n")
		format.PrintProgressComplete(fmt.Sprintf("Complete! Elapsed time: %s", formatDuration(elapsed)))
	}
}

// formatDuration formats a time.Duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	seconds := int(d.Seconds()) % 60
	minutes := int(d.Minutes()) % 60
	hours := int(d.Hours())

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// renderProgressBar renders the progress bar to stdout with modern styling
func renderProgressBar(percentage int) {
	// Adjust bar width for modern style
	actualBarWidth := 35
	filledWidth := int(float64(actualBarWidth) * float64(percentage) / 100.0)
	emptyWidth := actualBarWidth - filledWidth

	// Build the progress line
	var progressLine strings.Builder
	
	// Start with box border
	progressLine.WriteString("│ ")
	progressLine.WriteString(format.InfoColor.Sprint("►"))
	progressLine.WriteString(" Processing: ")
	
	// Progress bar
	progressLine.WriteString(format.ProgressBg.Sprint("["))
	
	// Filled portion
	for i := 0; i < filledWidth; i++ {
		progressLine.WriteString(format.ProgressFg.Sprint("█"))
	}
	
	// Empty portion
	for i := 0; i < emptyWidth; i++ {
		progressLine.WriteString(format.ProgressBg.Sprint("░"))
	}
	
	progressLine.WriteString(format.ProgressBg.Sprint("]"))
	
	// Percentage
	progressLine.WriteString(format.BaseHighlight.Sprintf(" %3d%%", percentage))
	
	// Elapsed time
	elapsed := time.Since(startTime)
	elapsedStr := formatDuration(elapsed)
	progressLine.WriteString(format.BaseDim.Sprintf(" • %s", elapsedStr))
	
	// Calculate visible length for padding
	// 2 (border) + 1 (icon) + 12 (" Processing: ") + 1 ([) + 35 (bar) + 1 (]) + 5 (percentage) + 3 (" • ") + len(elapsed)
	visibleLen := 2 + 1 + 12 + 1 + actualBarWidth + 1 + 5 + 3 + len(elapsedStr)
	padding := format.BoxWidth - visibleLen - 2
	
	if padding > 0 {
		progressLine.WriteString(strings.Repeat(" ", padding))
	}
	progressLine.WriteString(" │")
	
	// Print with carriage return to overwrite and clear to end of line
	fmt.Print("\r" + progressLine.String() + "\033[K")

	// Ensure the output is flushed immediately
	os.Stdout.Sync()
}

// ResetProgressBar resets the progress bar for a new operation
func ResetProgressBar() {
	once = sync.Once{}
	lastPercent = 0
	startTime = time.Time{}
}

// ParseProgressLine extracts percentage from mkvmerge progress output
func ParseProgressLine(line string) (int, bool) {
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
