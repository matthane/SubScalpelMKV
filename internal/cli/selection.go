package cli

import (
	"fmt"
	"strconv"
	"strings"

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
func ProcessSelectionAndExclusion(extractAll bool) (*SelectionResult, error) {
	result := &SelectionResult{}

	if !extractAll {
		selectionInput := AskTrackSelection()
		result.Selection = ParseTrackSelection(selectionInput)

		if len(result.Selection.LanguageCodes) == 0 && len(result.Selection.TrackNumbers) == 0 && len(result.Selection.FormatFilters) == 0 {
			// Empty input means accept all tracks - same as extractAll = true
			// Ask for exclusions when extracting all tracks
			exclusionInput := AskTrackExclusion()
			if exclusionInput != "" {
				exclusion := ParseTrackExclusion(exclusionInput)
				result.Selection.Exclusions = exclusion
				result.ExclusionFilter = convertExclusionToString(exclusion)
				result.Title = "Track Processing"
				result.Message = buildExclusionOnlyMessage(exclusion)
			} else {
				result.Title = "Track Processing"
				result.Message = "Extracting all subtitle tracks..."
			}
		} else {
			// Ask for exclusions after selection
			exclusionInput := AskTrackExclusion()
			if exclusionInput != "" {
				exclusion := ParseTrackExclusion(exclusionInput)
				result.Selection.Exclusions = exclusion
				result.ExclusionFilter = convertExclusionToString(exclusion)
			}

			// Convert to comma-separated string for processFile function
			result.LanguageFilter = convertSelectionToString(result.Selection)
			result.Title, result.Message = buildSelectionTitleAndMessage(result.Selection, result.Selection.Exclusions)
		}
	} else {
		// Ask for exclusions even when extracting all tracks
		exclusionInput := AskTrackExclusion()
		if exclusionInput != "" {
			exclusion := ParseTrackExclusion(exclusionInput)
			result.Selection.Exclusions = exclusion
			result.ExclusionFilter = convertExclusionToString(exclusion)
			result.Title = "Track Processing"
			result.Message = buildExclusionOnlyMessage(exclusion)
		} else {
			result.Title = "Track Processing"
			result.Message = "Extracting all subtitle tracks..."
		}
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