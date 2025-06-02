package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Custom RGB color helper function
func NewRGBColor(r, g, b int, attributes ...color.Attribute) *color.Color {
	attrs := []color.Attribute{color.Attribute(38), color.Attribute(2), color.Attribute(r), color.Attribute(g), color.Attribute(b)}
	attrs = append(attrs, attributes...)
	return color.New(attrs...)
}

// Color definitions for terminal output
var (
	// Title colors
	TitleWhiteColor = NewRGBColor(255, 255, 255, color.Bold) // Pure White
	TitleRedColor   = NewRGBColor(255, 215, 0, color.Bold)   // Gold

	// Header colors
	HeaderColor  = NewRGBColor(255, 215, 0, color.Bold) // Gold
	SectionColor = NewRGBColor(255, 215, 0, color.Bold) // Gold

	// Status colors
	SuccessColor = NewRGBColor(50, 205, 50, color.Bold) // Lime Green
	ErrorColor   = color.New(color.BgRed, color.FgWhite, color.Bold)
	WarningColor = NewRGBColor(255, 140, 0, color.Bold)   // Dark Orange
	InfoColor    = NewRGBColor(135, 206, 235, color.Bold) // Sky Blue

	// Track information colors
	TrackNumberColor = NewRGBColor(255, 127, 80, color.Bold) // Coral
	LanguageColor    = NewRGBColor(144, 238, 144)            // Light Green
	CodecColor       = NewRGBColor(0, 128, 128, color.Bold)  // Teal

	// Track attribute colors
	ForcedAttribute  = color.New(color.BgRed, color.FgWhite, color.Bold)
	DefaultAttribute = color.New(color.BgGreen, color.FgWhite, color.Bold)

	// Progress bar colors
	ProgressBarColor     = NewRGBColor(30, 144, 255, color.Bold) // Dodger Blue
	ProgressFillColor    = NewRGBColor(30, 144, 255, color.Bold) // Dodger Blue
	ProgressEmptyColor   = NewRGBColor(105, 105, 105)            // Dim Gray
	ProgressPercentColor = NewRGBColor(255, 215, 0, color.Bold)  // Gold

	// User interaction colors
	PromptColor = NewRGBColor(255, 255, 255, color.Bold) // Pure White
	InputColor  = NewRGBColor(135, 206, 235)             // Sky Blue

	// General purpose colors
	AccentColor    = NewRGBColor(255, 215, 0, color.Bold) // Gold
	HighlightColor = NewRGBColor(255, 215, 0, color.Bold) // Gold
	SubtleColor    = NewRGBColor(169, 169, 169)           // Dark Gray

	// Additional colors used by functions
	CriticalError      = color.New(color.BgHiRed, color.FgWhite, color.Bold)
	BrandColor         = NewRGBColor(255, 215, 0, color.Bold) // Gold
	VideoTrackColor    = NewRGBColor(255, 165, 0, color.Bold) // Orange
	AudioTrackColor    = NewRGBColor(128, 0, 128, color.Bold) // Purple
	SubtitleTrackColor = NewRGBColor(128, 0, 128, color.Bold) // Purple
	ChapterTrackColor  = NewRGBColor(255, 20, 147)            // Deep Pink
	ProcessingStatus   = color.New(color.BgBlue, color.FgWhite)
	ImportantNotice    = color.New(color.BgYellow, color.FgBlack, color.Bold)
)

// PrintTitle prints the main application title
func PrintTitle() {
	fmt.Printf("🎞️🗡️ %s%s\n",
		TitleWhiteColor.Sprint("SubScalpel"),
		TitleRedColor.Sprint("MKV"))
	TitleWhiteColor.Println("===================")
}

// PrintSection prints a section header with decorative formatting
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

// PrintSuccess prints a success message with checkmark
func PrintSuccess(message string) {
	SuccessColor.Printf("✓  %s\n", message)
}

// PrintError prints an error message with highlighted background
func PrintError(message string) {
	ErrorColor.Printf(" ✗ %s ", message)
	fmt.Println()
}

// PrintCriticalError prints a critical error with prominent background
func PrintCriticalError(message string) {
	CriticalError.Printf(" ❌ %s ", message)
	fmt.Println()
}

// PrintWarning prints a warning message with warning icon
func PrintWarning(message string) {
	WarningColor.Printf("⚠  %s\n", message)
}

// PrintInfo prints an informational message with info icon
func PrintInfo(message string) {
	InfoColor.Printf("ℹ  %s\n", message)
}

// PrintStep prints a numbered step message
func PrintStep(step int, message string) {
	AccentColor.Printf("Step %d: ", step)
	BrandColor.Printf("%s\n", message)
}

// PrintTrackInfo prints formatted track information
func PrintTrackInfo(trackNum int, language, trackName, codecType string, forced, defaultTrack bool) {
	PrintTrackInfoWithType(trackNum, "", language, trackName, codecType, forced, defaultTrack)
}

// PrintTrackInfoWithType prints formatted track information with type-specific colors
func PrintTrackInfoWithType(trackNum int, trackType, language, trackName, codecType string, forced, defaultTrack bool) {
	var parts []string
	var trackColor *color.Color

	// Choose color based on track type
	switch strings.ToLower(trackType) {
	case "video":
		trackColor = VideoTrackColor
	case "audio":
		trackColor = AudioTrackColor
	case "subtitle", "subtitles":
		trackColor = SubtitleTrackColor
	case "chapter", "chapters":
		trackColor = ChapterTrackColor
	default:
		trackColor = TrackNumberColor
	}

	trackInfo := fmt.Sprintf("Track ID %s: %s",
		trackColor.Sprint(trackNum),
		LanguageColor.Sprint(language))

	if trackType != "" {
		trackInfo += fmt.Sprintf(" (%s)", trackType)
	}
	parts = append(parts, trackInfo)

	if trackName != "" {
		parts = append(parts, fmt.Sprintf("(%s)", AccentColor.Sprint(trackName)))
	}

	// Track attributes with background highlighting
	var attributes []string
	if forced {
		attributes = append(attributes, ForcedAttribute.Sprint(" FORCED "))
	}
	if defaultTrack {
		attributes = append(attributes, DefaultAttribute.Sprint(" DEFAULT "))
	}
	if codecType != "" {
		attributes = append(attributes, CodecColor.Sprintf("[%s]", codecType))
	}

	if len(attributes) > 0 {
		parts = append(parts, strings.Join(attributes, " "))
	}

	fmt.Printf("  %s\n", strings.Join(parts, " "))
}

// PrintPrompt prints a user prompt
func PrintPrompt(message string) {
	PromptColor.Print(message)
}

// PrintFilter prints filter information with description
func PrintFilter(filterType string, values interface{}) {
	InfoColor.Printf("%s: ", filterType)
	HighlightColor.Printf("%v", values)
	SubtleColor.Println(" (only muxing and extracting matching tracks)")
}

// PrintProgressWithPercentage prints file processing progress
func PrintProgressWithPercentage(filename string, percentage int) {
	ProcessingStatus.Printf(" 🎬 Processing ")
	fmt.Print(" ")
	BrandColor.Printf("%s ", filename)
	HighlightColor.Printf("(%d%% complete)\n", percentage)
}

// PrintProgressComplete prints completion message with checkmark
func PrintProgressComplete(message string) {
	SuccessColor.Printf("✅ %s\n", message)
}

// PrintUsageSection prints a help section with title
func PrintUsageSection(title, content string) {
	fmt.Println()
	HeaderColor.Printf("%s:\n", title)
	fmt.Print(content)
}

// PrintExample prints a command example
func PrintExample(command string) {
	InputColor.Printf("  %s\n", command)
}

// PrintImportantNotice prints a highlighted notice
func PrintImportantNotice(message string) {
	ImportantNotice.Printf(" ⚠ %s ", message)
	fmt.Println()
}

// PrintPlain prints text without color formatting
func PrintPlain(message string) {
	fmt.Println(message)
}
