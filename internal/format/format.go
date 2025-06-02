package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Color definitions for different types of output
var (
	// Header colors
	TitleColor   = color.New(color.FgCyan, color.Bold)
	HeaderColor  = color.New(color.FgYellow, color.Bold)
	SectionColor = color.New(color.FgBlue, color.Bold)

	// Status colors
	SuccessColor = color.New(color.FgGreen, color.Bold)
	ErrorColor   = color.New(color.FgRed, color.Bold)
	WarningColor = color.New(color.FgYellow)
	InfoColor    = color.New(color.FgCyan)

	// Track information colors
	TrackNumberColor = color.New(color.FgMagenta, color.Bold)
	LanguageColor    = color.New(color.FgGreen)
	CodecColor       = color.New(color.FgBlue)
	AttributeColor   = color.New(color.FgYellow)

	// Progress colors
	ProgressColor        = color.New(color.FgBlue, color.Bold)
	PercentageColor      = color.New(color.FgYellow, color.Bold)
	ProgressBarColor     = color.New(color.FgBlue, color.Bold)
	ProgressFillColor    = color.New(color.FgBlue, color.Bold)
	ProgressEmptyColor   = color.New(color.FgHiBlack)
	ProgressPercentColor = color.New(color.FgYellow, color.Bold)

	// Prompt colors
	PromptColor = color.New(color.FgWhite, color.Bold)
	InputColor  = color.New(color.FgCyan)
)

// PrintTitle prints the main application title with formatting
func PrintTitle() {
	TitleColor.Println("🎞️🗡️ SubScalpelMKV")
	HeaderColor.Println("===================")
}

// PrintSection prints a section header with formatting
func PrintSection(title string) {
	fmt.Println()
	SectionColor.Printf("▶ %s\n", title)
	fmt.Println(strings.Repeat("─", len(title)+2))
}

// PrintSubSection prints a subsection header
func PrintSubSection(title string) {
	fmt.Println()
	HeaderColor.Printf("● %s\n", title)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	SuccessColor.Printf("✓  %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	ErrorColor.Printf("✗  %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	WarningColor.Printf("⚠  %s\n", message)
}

// PrintInfo prints an informational message
func PrintInfo(message string) {
	InfoColor.Printf("ℹ  %s\n", message)
}

// PrintStep prints a step message
func PrintStep(step int, message string) {
	SectionColor.Printf("Step %d: %s\n", step, message)
}

// PrintTrackInfo prints formatted track information
func PrintTrackInfo(trackNum int, language, trackName, codecType string, forced, defaultTrack bool) {
	var parts []string

	// Track number and language
	trackInfo := fmt.Sprintf("Track %s: %s",
		TrackNumberColor.Sprint(trackNum),
		LanguageColor.Sprint(language))
	parts = append(parts, trackInfo)

	// Track name if available
	if trackName != "" {
		parts = append(parts, fmt.Sprintf("(%s)", trackName))
	}

	// Attributes
	var attributes []string
	if forced {
		attributes = append(attributes, AttributeColor.Sprint("FORCED"))
	}
	if defaultTrack {
		attributes = append(attributes, AttributeColor.Sprint("DEFAULT"))
	}
	if codecType != "" {
		attributes = append(attributes, CodecColor.Sprintf("[%s]", codecType))
	}

	if len(attributes) > 0 {
		parts = append(parts, strings.Join(attributes, " "))
	}

	fmt.Printf("  %s\n", strings.Join(parts, " "))
}

// PrintPrompt prints a user prompt with formatting
func PrintPrompt(message string) {
	PromptColor.Print(message)
}

// PrintFilter prints filter information
func PrintFilter(filterType string, values interface{}) {
	InfoColor.Printf("%s: %v (only muxing and extracting matching tracks)\n", filterType, values)
}

// PrintProgress prints progress information with colors
func PrintProgress(message string) {
	ProgressColor.Printf("🎬 %s\n", message)
}

// PrintProgressComplete prints completion message
func PrintProgressComplete(message string) {
	SuccessColor.Printf("✅ %s\n", message)
}

// PrintUsageSection prints a usage section with proper formatting
func PrintUsageSection(title, content string) {
	fmt.Println()
	HeaderColor.Printf("%s:\n", title)
	fmt.Print(content)
}

// PrintExample prints an example with formatting
func PrintExample(command string) {
	InputColor.Printf("  %s\n", command)
}

// PrintNoColor versions for when color should be disabled
func PrintPlain(message string) {
	fmt.Println(message)
}

// FormatDuration formats duration with color
func FormatDuration(duration string) string {
	return SuccessColor.Sprint(duration)
}
