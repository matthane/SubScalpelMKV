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

// Modern color palette inspired by btop and other modern terminal apps
var (
	// Base colors - sophisticated and muted
	BaseFg        = NewRGBColor(220, 220, 220)              // Light gray for main text
	BaseDim       = NewRGBColor(110, 110, 120)              // Dimmed text
	BaseAccent    = NewRGBColor(140, 170, 220)              // Soft blue accent
	BaseHighlight = NewRGBColor(255, 255, 255, color.Bold)  // Pure white for emphasis

	// UI elements
	BorderColor  = NewRGBColor(80, 85, 95)                  // Dark gray borders
	HeaderBg     = NewRGBColor(50, 55, 65)                  // Dark header background
	SectionColor = NewRGBColor(180, 190, 210)                // Light blue-gray for sections

	// Status colors - softer, more professional
	SuccessColor = NewRGBColor(115, 190, 120)               // Soft green
	ErrorColor   = NewRGBColor(215, 95, 95)                 // Soft red
	WarningColor = NewRGBColor(220, 180, 90)                // Soft amber
	InfoColor    = NewRGBColor(130, 160, 210)               // Soft blue

	// Track type colors - more muted and harmonious
	VideoTrackColor    = NewRGBColor(180, 140, 200)         // Soft purple
	AudioTrackColor    = NewRGBColor(140, 180, 200)         // Soft cyan
	SubtitleTrackColor = NewRGBColor(200, 180, 140)         // Soft yellow
	ChapterTrackColor  = NewRGBColor(200, 140, 140)         // Soft pink

	// Track information colors
	TrackNumberColor = BaseHighlight                        // White for track numbers
	LanguageColor    = BaseFg                               // Standard text for language
	CodecColor       = BaseDim                              // Dimmed for codec info

	// Track attribute colors - modern style
	ForcedAttribute  = WarningColor                         // Use warning color for forced
	DefaultAttribute = SuccessColor                         // Use success color for default

	// Progress colors - modern gradient effect
	ProgressFg   = NewRGBColor(100, 180, 240)              // Bright blue
	ProgressBg   = NewRGBColor(50, 55, 65)                 // Dark background
	ProgressText = NewRGBColor(200, 210, 220)              // Light text

	// User interaction colors
	PromptColor = BaseHighlight                             // White for prompts
	InputColor  = BaseAccent                                // Blue accent for input

)


// PrintTitleWithVersion prints the main application title with version number
func PrintTitleWithVersion(version string) {
	titleWidth := 30 // Fixed width for title box
	
	// Top border with title
	title := "SubScalpelMKV"
	titleLen := len(title)
	dashesBeforeTitle := 1
	dashesAfterTitle := titleWidth - titleLen - dashesBeforeTitle - 2 // -2 for spaces around title
	
	BaseAccent.Print("┌")
	BaseAccent.Print(strings.Repeat("─", dashesBeforeTitle))
	BaseAccent.Print(" ")
	BaseHighlight.Print("SubScalpel")
	BaseFg.Print("MKV")
	BaseAccent.Print(" ")
	BaseAccent.Print(strings.Repeat("─", dashesAfterTitle))
	BaseAccent.Println("┐")
	
	// Middle line
	subtitle := "Extract MKV Subtitles"
	if version != "" {
		subtitle = fmt.Sprintf("Extract MKV Subtitles v%s", version)
	}
	subtitleLen := len(subtitle)
	padding := titleWidth - subtitleLen - 2 // -2 for "│ " at start
	
	BaseAccent.Print("│ ")
	BaseDim.Print(subtitle)
	fmt.Print(strings.Repeat(" ", padding))
	BaseAccent.Println(" │")
	
	// Bottom border
	BaseAccent.Print("└")
	BaseAccent.Print(strings.Repeat("─", titleWidth))
	BaseAccent.Println("┘")
}

// Box width constant for consistent sizing
const BoxWidth = 60

// PrintSection prints a section header with modern box drawing
func PrintSection(title string) {
	fmt.Println()
	titlePadded := fmt.Sprintf(" %s ", title)
	titleLen := len(titlePadded)
	leftPad := (BoxWidth - titleLen) / 2
	rightPad := BoxWidth - titleLen - leftPad
	
	BorderColor.Print("╭")
	BorderColor.Print(strings.Repeat("─", leftPad))
	SectionColor.Print(titlePadded)
	BorderColor.Print(strings.Repeat("─", rightPad))
	BorderColor.Println("╮")
}

// PrintSubSection prints a subsection header
func PrintSubSection(title string) {
	fmt.Println()
	SectionColor.Printf("● %s", title)
}

// PrintSuccess prints a success message with modern styling
func PrintSuccess(message string) {
	SuccessColor.Print("  ✓ ")
	BaseFg.Println(message)
}

// PrintError prints an error message with modern styling
func PrintError(message string) {
	ErrorColor.Print("  ✗ ")
	BaseFg.Println(message)
}


// PrintWarning prints a warning message with modern styling
func PrintWarning(message string) {
	WarningColor.Print("  ⚡ ")
	BaseFg.Println(message)
}

// PrintInfo prints an informational message with modern styling
func PrintInfo(message string) {
	InfoColor.Print("  ◆ ")
	BaseFg.Println(message)
}

// PrintStep prints a numbered step message with modern styling
func PrintStep(step int, message string) {
	fmt.Print("  ")
	InfoColor.Print("►")
	fmt.Print(" ")
	BaseDim.Printf("Step %d:", step)
	fmt.Print(" ")
	BaseFg.Println(message)
}

// PrintTrackInfoWithLanguageName prints formatted track information with full language name
func PrintTrackInfoWithLanguageName(trackNum int, language, languageName, trackName, codecType string, forced, defaultTrack bool) {
	// Use white for the track indicator
	trackColor := BaseHighlight

	// First line: Track info
	// Print each part separately to avoid ANSI code length issues
	BorderColor.Print("│ ")
	trackColor.Print("▪")
	fmt.Print(" ")
	BaseFg.Print("Track ")
	BaseHighlight.Print(trackNum)
	BaseDim.Print(" • ")
	BaseFg.Print(language)
	
	// Calculate visible content length for first line
	contentLen := 2 + 2 + 6 + len(fmt.Sprint(trackNum)) + 3 + len(language) // "│ " + "▪ " + "Track " + num + " • " + lang
	
	// Add full language name if provided
	if languageName != "" && languageName != language {
		BaseDim.Print(" (")
		BaseAccent.Print(languageName)
		BaseDim.Print(")")
		contentLen += 3 + len(languageName) // " (" + name + ")"
	}
	
	if trackName != "" {
		BaseDim.Print(" • ")
		BaseAccent.Print(trackName)
		contentLen += 3 + len(trackName)
	}
	
	// Add padding and close the line
	padding := BoxWidth - contentLen // No need to subtract 1 for track line
	if padding > 0 {
		fmt.Print(strings.Repeat(" ", padding))
	}
	BorderColor.Println(" │")
	
	// Second line: Attributes (if any)
	if forced || defaultTrack || codecType != "" {
		BorderColor.Print("│   ")
		attrLen := 3 // "│   "
		
		if defaultTrack {
			DefaultAttribute.Print("◉ DEFAULT")
			attrLen += 9
			if forced || codecType != "" {
				fmt.Print("  ")
				attrLen += 2
			}
		}
		
		if forced {
			ForcedAttribute.Print("◉ FORCED")
			attrLen += 8
			if codecType != "" {
				fmt.Print("  ")
				attrLen += 2
			}
		}
		
		if codecType != "" {
			CodecColor.Print(codecType)
			attrLen += len(codecType)
		}
		
		// Add padding and close the line
		attrPadding := BoxWidth - attrLen - 1 // -1 for space before closing border
		if attrPadding > 0 {
			fmt.Print(strings.Repeat(" ", attrPadding))
		}
		BorderColor.Println(" │")
	}
}

// PrintPrompt prints a user prompt with modern styling
func PrintPrompt(message string) {
	fmt.Print("  ")
	PromptColor.Print("▸ ")
	BaseFg.Print(message)
}

// PrintPromptWithPlaceholder prints a user prompt with placeholder text
func PrintPromptWithPlaceholder(message, placeholder string) {
	fmt.Print("  ")
	PromptColor.Print("▸ ")
	BaseFg.Print(message)
	if placeholder != "" {
		BaseDim.Printf("%s ", placeholder)
	}
}

// PrintFilter prints filter information with modern styling
func PrintFilter(filterType string, values interface{}) {
	fmt.Print("  ")
	BaseDim.Printf("%s: ", filterType)
	BaseHighlight.Printf("%v", values)
	BaseDim.Println(" (filtered)")
}


// PrintUsageSection prints a help section with title
func PrintUsageSection(title, content string) {
	fmt.Println()
	SectionColor.Printf("%s:", title)
	fmt.Print(content)
}

// PrintExample prints a command example
func PrintExample(command string) {
	InputColor.Printf("  %s\n", command)
}


// DrawBoxBottom draws the bottom of a box with modern styling
func DrawBoxBottom(width int) {
	BorderColor.Print("╰")
	BorderColor.Print(strings.Repeat("─", width))
	BorderColor.Println("╯")
}

// DrawSeparator draws a separator line inside a box
func DrawSeparator(width int) {
	BorderColor.Print("│ ")
	BaseDim.Print(strings.Repeat("·", width-2))
	BorderColor.Println(" │")
}
