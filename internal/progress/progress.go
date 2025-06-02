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

// Default theme with colors
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

	// Only update if percentage has increased
	if percentage > lastPercent {
		renderProgressBar(percentage)
		lastPercent = percentage
	}

	// If we've reached 100%, show completion message
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

// renderProgressBar renders the progress bar to stdout
func renderProgressBar(percentage int) {
	// Calculate filled and empty portions
	filledWidth := int(float64(barWidth) * float64(percentage) / 100.0)
	emptyWidth := barWidth - filledWidth

	// Build the progress bar string
	var bar strings.Builder

	// Start character
	bar.WriteString(format.ProgressBarColor.Sprint(defaultTheme.BarStart))

	// Filled portion
	for i := 0; i < filledWidth; i++ {
		bar.WriteString(format.ProgressFillColor.Sprint(defaultTheme.Saucer))
	}

	// Empty portion
	for i := 0; i < emptyWidth; i++ {
		bar.WriteString(format.ProgressEmptyColor.Sprint(defaultTheme.SaucerPadding))
	}

	// End character
	bar.WriteString(format.ProgressBarColor.Sprint(defaultTheme.BarEnd))

	// Print the progress bar with percentage
	// Use a consistent format and ensure we overwrite the entire line
	progressLine := fmt.Sprintf("%s %s", bar.String(), format.ProgressPercentColor.Sprintf("%3d%%", percentage))

	// Pad with spaces to ensure we overwrite any previous longer content
	const minLineLength = 120
	if len(progressLine) < minLineLength {
		progressLine += strings.Repeat(" ", minLineLength-len(progressLine))
	}

	fmt.Printf("\r%s", progressLine)

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
