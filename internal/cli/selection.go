package cli

import (
	"fmt"
	"strconv"
	"strings"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/model"
)

// SelectionResult contains the processed selection and exclusion filters
type SelectionResult struct {
	LanguageFilter  string
	ExclusionFilter string
	Selection       model.TrackSelection
	Message         string
	Title           string
}

// ProcessSelectionAndExclusion handles the common logic for processing track selections and exclusions
func ProcessSelectionAndExclusion(extractAll bool, availableTracks []int) (*SelectionResult, error) {
	result := &SelectionResult{}

	if !extractAll {
		// Get selection with validation and retry
		var selectionInput string
		var validSelection bool
		for !validSelection {
			selectionInput = AskTrackSelection()
			var invalidItems []string
			result.Selection, invalidItems = ParseTrackSelectionWithValidation(selectionInput, availableTracks)
			
			if len(invalidItems) > 0 {
				// Show warning and ask to retry
				for _, item := range invalidItems {
					format.PrintWarning(fmt.Sprintf("Unknown language code, format, or invalid track ID '%s'", item))
				}
				fmt.Println() // Add spacing
				continue
			}
			validSelection = true
		}

		if len(result.Selection.LanguageCodes) == 0 && len(result.Selection.TrackNumbers) == 0 && len(result.Selection.FormatFilters) == 0 {
			// Empty input means accept all tracks - same as extractAll = true
			// Ask for exclusions when extracting all tracks
			var exclusionInput string
			var validExclusion bool
			for !validExclusion {
				exclusionInput = AskTrackExclusion()
				if exclusionInput != "" {
					var invalidItems []string
					var exclusion model.TrackExclusion
					exclusion, invalidItems = ParseTrackExclusionWithValidation(exclusionInput, availableTracks)
					
					if len(invalidItems) > 0 {
						// Show warning and ask to retry
						for _, item := range invalidItems {
							format.PrintWarning(fmt.Sprintf("Unknown exclusion language code, format, or invalid track ID '%s'", item))
						}
						fmt.Println() // Add spacing
						continue
					}
					
					result.Selection.Exclusions = exclusion
					result.ExclusionFilter = convertExclusionToString(exclusion)
					result.Title = "Track Processing"
					result.Message = buildExclusionOnlyMessage(exclusion)
				} else {
					result.Title = "Track Processing"
					result.Message = "Extracting all subtitle tracks..."
				}
				validExclusion = true
			}
		} else {
			// Ask for exclusions after selection
			var exclusionInput string
			var validExclusion bool
			for !validExclusion {
				exclusionInput = AskTrackExclusion()
				if exclusionInput != "" {
					var invalidItems []string
					var exclusion model.TrackExclusion
					exclusion, invalidItems = ParseTrackExclusionWithValidation(exclusionInput, availableTracks)
					
					if len(invalidItems) > 0 {
						// Show warning and ask to retry
						for _, item := range invalidItems {
							format.PrintWarning(fmt.Sprintf("Unknown exclusion language code, format, or invalid track ID '%s'", item))
						}
						fmt.Println() // Add spacing
						continue
					}
					
					result.Selection.Exclusions = exclusion
					result.ExclusionFilter = convertExclusionToString(exclusion)
				}
				validExclusion = true
			}

			// Convert to comma-separated string for processFile function
			result.LanguageFilter = convertSelectionToString(result.Selection)
			result.Title, result.Message = buildSelectionTitleAndMessage(result.Selection, result.Selection.Exclusions)
		}
	} else {
		// When extracting all tracks, don't ask for exclusions - just extract everything
		result.Title = "Track Processing"
		result.Message = "Extracting all subtitle tracks..."
	}

	return result, nil
}

// ProcessSelectionForBatch handles selection without interactive prompts (for batch mode)
func ProcessSelectionForBatch(selection model.TrackSelection, exclusion model.TrackExclusion) *SelectionResult {
	result := &SelectionResult{
		Selection: selection,
	}
	result.Selection.Exclusions = exclusion

	if len(selection.LanguageCodes) > 0 || len(selection.TrackNumbers) > 0 || len(selection.FormatFilters) > 0 {
		result.LanguageFilter = convertSelectionToString(selection)
	}

	if len(exclusion.LanguageCodes) > 0 || len(exclusion.TrackNumbers) > 0 || len(exclusion.FormatFilters) > 0 {
		result.ExclusionFilter = convertExclusionToString(exclusion)
	}

	if result.LanguageFilter != "" {
		result.Title, result.Message = buildSelectionTitleAndMessage(selection, exclusion)
	} else if result.ExclusionFilter != "" {
		result.Title = "Track Processing"
		result.Message = buildExclusionOnlyMessage(exclusion)
	}

	return result
}

// convertSelectionToString converts a TrackSelection to a comma-separated string
func convertSelectionToString(selection model.TrackSelection) string {
	var filterParts []string
	filterParts = append(filterParts, selection.LanguageCodes...)
	for _, trackNum := range selection.TrackNumbers {
		filterParts = append(filterParts, strconv.Itoa(trackNum))
	}
	filterParts = append(filterParts, selection.FormatFilters...)
	return strings.Join(filterParts, ",")
}

// convertExclusionToString converts a TrackExclusion to a comma-separated string
func convertExclusionToString(exclusion model.TrackExclusion) string {
	var exclusionParts []string
	exclusionParts = append(exclusionParts, exclusion.LanguageCodes...)
	for _, trackNum := range exclusion.TrackNumbers {
		exclusionParts = append(exclusionParts, strconv.Itoa(trackNum))
	}
	exclusionParts = append(exclusionParts, exclusion.FormatFilters...)
	return strings.Join(exclusionParts, ",")
}

// buildSelectionTitleAndMessage builds a user-friendly title and message for the selection and exclusion
func buildSelectionTitleAndMessage(selection model.TrackSelection, exclusion model.TrackExclusion) (string, string) {
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

	if len(messageParts) == 0 {
		return "", ""
	}

	baseMessage := fmt.Sprintf("Extracting tracks for %s", strings.Join(messageParts, ", "))

	// Add exclusion info if present
	if len(exclusion.LanguageCodes) > 0 || len(exclusion.TrackNumbers) > 0 || len(exclusion.FormatFilters) > 0 {
		var exclusionMsgParts []string
		if len(exclusion.LanguageCodes) > 0 {
			exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("languages: %s", strings.Join(exclusion.LanguageCodes, ",")))
		}
		if len(exclusion.TrackNumbers) > 0 {
			exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("track IDs: %v", exclusion.TrackNumbers))
		}
		if len(exclusion.FormatFilters) > 0 {
			exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("formats: %s", strings.Join(exclusion.FormatFilters, ",")))
		}

		if len(exclusionMsgParts) > 0 {
			baseMessage = fmt.Sprintf("%s, excluding %s", baseMessage, strings.Join(exclusionMsgParts, ", "))
		}
	}

	return "Track Processing", baseMessage
}

// buildExclusionOnlyMessage builds a message when only exclusions are specified
func buildExclusionOnlyMessage(exclusion model.TrackExclusion) string {
	var exclusionMsgParts []string
	if len(exclusion.LanguageCodes) > 0 {
		exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("languages: %s", strings.Join(exclusion.LanguageCodes, ",")))
	}
	if len(exclusion.TrackNumbers) > 0 {
		exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("track IDs: %v", exclusion.TrackNumbers))
	}
	if len(exclusion.FormatFilters) > 0 {
		exclusionMsgParts = append(exclusionMsgParts, fmt.Sprintf("formats: %s", strings.Join(exclusion.FormatFilters, ",")))
	}

	if len(exclusionMsgParts) > 0 {
		return fmt.Sprintf("Extracting all tracks except %s", strings.Join(exclusionMsgParts, ", "))
	}
	return "Extracting all subtitle tracks..."
}

// ParseTrackSelectionWithValidation parses track selection input and returns invalid items
func ParseTrackSelectionWithValidation(input string, availableTracks []int) (model.TrackSelection, []string) {
	selection := model.TrackSelection{
		LanguageCodes: []string{},
		TrackNumbers:  []int{},
		FormatFilters: []string{},
		Exclusions:    model.TrackExclusion{},
	}
	
	var invalidItems []string

	if input == "" {
		return selection, invalidItems
	}

	items := strings.Split(input, ",")

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Try to parse as track number first
		if trackNum, err := strconv.Atoi(item); err == nil {
			// Check if track number is valid
			isValidTrack := false
			for _, validTrack := range availableTracks {
				if trackNum == validTrack {
					isValidTrack = true
					break
				}
			}
			if isValidTrack {
				selection.TrackNumbers = append(selection.TrackNumbers, trackNum)
				continue
			} else {
				invalidItems = append(invalidItems, item)
				continue
			}
		}

		// Try to parse as language code
		isValidLanguage := false
		if len(item) == 2 {
			_, isValidLanguage = model.LanguageCodeMapping[strings.ToLower(item)]
		} else if len(item) == 3 {
			for _, threeLetter := range model.LanguageCodeMapping {
				if strings.EqualFold(item, threeLetter) {
					isValidLanguage = true
					break
				}
			}
		}

		if isValidLanguage {
			selection.LanguageCodes = append(selection.LanguageCodes, item)
			continue
		}

		// Try to parse as subtitle format filter
		isValidFormat := false
		lowerItem := strings.ToLower(item)
		for _, ext := range model.SubtitleExtensionByCodec {
			if lowerItem == ext {
				isValidFormat = true
				break
			}
		}

		if isValidFormat {
			selection.FormatFilters = append(selection.FormatFilters, lowerItem)
		} else {
			invalidItems = append(invalidItems, item)
		}
	}

	return selection, invalidItems
}

// ParseTrackExclusionWithValidation parses track exclusion input and returns invalid items
func ParseTrackExclusionWithValidation(input string, availableTracks []int) (model.TrackExclusion, []string) {
	exclusion := model.TrackExclusion{
		LanguageCodes: []string{},
		TrackNumbers:  []int{},
		FormatFilters: []string{},
	}
	
	var invalidItems []string

	if input == "" {
		return exclusion, invalidItems
	}

	items := strings.Split(input, ",")

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Try to parse as track number first
		if trackNum, err := strconv.Atoi(item); err == nil {
			// Check if track number is valid
			isValidTrack := false
			for _, validTrack := range availableTracks {
				if trackNum == validTrack {
					isValidTrack = true
					break
				}
			}
			if isValidTrack {
				exclusion.TrackNumbers = append(exclusion.TrackNumbers, trackNum)
				continue
			} else {
				invalidItems = append(invalidItems, item)
				continue
			}
		}

		// Try to parse as language code
		isValidLanguage := false
		if len(item) == 2 {
			_, isValidLanguage = model.LanguageCodeMapping[strings.ToLower(item)]
		} else if len(item) == 3 {
			for _, threeLetter := range model.LanguageCodeMapping {
				if strings.EqualFold(item, threeLetter) {
					isValidLanguage = true
					break
				}
			}
		}

		if isValidLanguage {
			exclusion.LanguageCodes = append(exclusion.LanguageCodes, item)
			continue
		}

		// Try to parse as subtitle format filter
		isValidFormat := false
		lowerItem := strings.ToLower(item)
		for _, ext := range model.SubtitleExtensionByCodec {
			if lowerItem == ext {
				isValidFormat = true
				break
			}
		}

		if isValidFormat {
			exclusion.FormatFilters = append(exclusion.FormatFilters, lowerItem)
		} else {
			invalidItems = append(invalidItems, item)
		}
	}

	return exclusion, invalidItems
}