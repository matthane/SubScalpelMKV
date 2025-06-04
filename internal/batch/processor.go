package batch

import (
	"fmt"
	"path/filepath"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/mkv"
	"subscalpelmkv/internal/model"
	"subscalpelmkv/internal/util"
)

// ProcessFileFunc is the function signature for processing a single file
type ProcessFileFunc func(inputFileName, languageFilter, exclusionFilter string, showFilterMessage bool, outputConfig model.OutputConfig, dryRun bool) error

// Processor handles batch processing of MKV files
type Processor struct {
	Files        []string
	OutputConfig model.OutputConfig
	DryRun       bool
}

// ProcessingResult contains the results of batch processing
type ProcessingResult struct {
	SuccessCount int
	ErrorCount   int
	TotalFiles   int
}

// NewProcessor creates a new batch processor
func NewProcessor(files []string, outputConfig model.OutputConfig, dryRun bool) *Processor {
	return &Processor{
		Files:        files,
		OutputConfig: outputConfig,
		DryRun:       dryRun,
	}
}

// Process executes the batch processing with the given processing function
func (p *Processor) Process(processFunc ProcessFileFunc, languageFilter, exclusionFilter string) (*ProcessingResult, error) {
	result := &ProcessingResult{
		TotalFiles: len(p.Files),
	}

	for i, file := range p.Files {
		format.PrintSubSection(fmt.Sprintf("Processing file %d/%d: %s", i+1, len(p.Files), filepath.Base(file)))
		
		err := processFunc(file, languageFilter, exclusionFilter, false, p.OutputConfig, p.DryRun)
		if err != nil {
			format.PrintError(fmt.Sprintf("Failed to process %s: %v", file, err))
			result.ErrorCount++
		} else {
			format.PrintSuccess(fmt.Sprintf("Successfully processed %s", filepath.Base(file)))
			result.SuccessCount++
		}
		
		// Add spacing between files except for the last one
		if i < len(p.Files)-1 {
			fmt.Println()
		}
	}

	return result, nil
}

// PrintSummary displays the batch processing summary
func (p *Processor) PrintSummary(result *ProcessingResult) {
	fmt.Println()
	format.PrintSubSection("Batch Processing Summary")
	format.PrintInfo(fmt.Sprintf("Total files: %d", result.TotalFiles))
	format.PrintSuccess(fmt.Sprintf("Successfully processed: %d", result.SuccessCount))
	if result.ErrorCount > 0 {
		format.PrintError(fmt.Sprintf("Failed to process: %d", result.ErrorCount))
	}
}

// AnalyzeFiles analyzes a list of files and returns their information
func AnalyzeFiles(files []string) []model.BatchFileInfo {
	var batchFileInfos []model.BatchFileInfo
	
	for _, file := range files {
		fileInfo := model.BatchFileInfo{
			FileName: filepath.Base(file),
			FilePath: file,
		}
		
		// Try to get track information for this file
		mkvInfo, err := mkv.GetTrackInfo(file)
		if err != nil {
			fileInfo.HasError = true
			fileInfo.ErrorMessage = fmt.Sprintf("Failed to analyze: %v", err)
		} else {
			// Count subtitle tracks and gather language codes and formats
			languageSet := make(map[string]bool)
			formatSet := make(map[string]bool)
			
			for _, track := range mkvInfo.Tracks {
				if track.Type == "subtitles" {
					fileInfo.SubtitleCount++
					
					// Collect language codes
					if track.Properties.Language != "" {
						languageSet[track.Properties.Language] = true
					}
					
					// Collect formats
					if ext, exists := model.SubtitleExtensionByCodec[track.Properties.CodecId]; exists {
						formatSet[ext] = true
					}
				}
			}
			
			// Convert sets to slices
			for lang := range languageSet {
				fileInfo.LanguageCodes = append(fileInfo.LanguageCodes, lang)
			}
			for format := range formatSet {
				fileInfo.SubtitleFormats = append(fileInfo.SubtitleFormats, format)
			}
		}
		
		batchFileInfos = append(batchFileInfos, fileInfo)
	}
	
	return batchFileInfos
}

// FilterValidFiles returns only the files that were successfully analyzed
func FilterValidFiles(fileInfos []model.BatchFileInfo) []string {
	var validFiles []string
	for _, fileInfo := range fileInfos {
		if !fileInfo.HasError {
			validFiles = append(validFiles, fileInfo.FilePath)
		}
	}
	return validFiles
}

// FilterMKVFiles filters a list of files to only include MKV files
func FilterMKVFiles(files []string) []string {
	var mkvFiles []string
	for _, file := range files {
		if util.IsMKVFile(file) {
			mkvFiles = append(mkvFiles, file)
		}
	}
	return mkvFiles
}