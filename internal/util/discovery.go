package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/model"
)

// DiscoverMKVFiles finds MKV files from command line arguments
// It handles individual files, multiple files, and directories
func DiscoverMKVFiles(args []string) ([]string, error) {
	var validMKVFiles []string
	var directories []string
	
	// First, check each argument
	for _, arg := range args {
		if info, err := os.Stat(arg); err == nil {
			if info.IsDir() {
				directories = append(directories, arg)
			} else if IsMKVFile(arg) {
				validMKVFiles = append(validMKVFiles, arg)
			}
		}
	}
	
	// Process any directories to find MKV files
	for _, dir := range directories {
		format.PrintInfo(fmt.Sprintf("Scanning directory: %s", dir))
		files, err := FindMKVFilesInDirectory(dir)
		if err != nil {
			format.PrintWarning(fmt.Sprintf("Error scanning directory %s: %v", dir, err))
			continue
		}
		validMKVFiles = append(validMKVFiles, files...)
	}
	
	return validMKVFiles, nil
}

// ValidateAndFilterMKVFiles validates a list of file paths and returns only valid MKV files
func ValidateAndFilterMKVFiles(files []string) ([]string, error) {
	var mkvFiles []string
	
	for _, file := range files {
		if info, err := os.Stat(file); err == nil && !info.IsDir() && IsMKVFile(file) {
			mkvFiles = append(mkvFiles, file)
		}
	}
	
	if len(mkvFiles) == 0 {
		return nil, fmt.Errorf("no MKV files found")
	}
	
	return mkvFiles, nil
}

// BuildOutputConfig creates an OutputConfig with special handling for batch mode
func BuildOutputConfig(outputDir, outputTemplate string, hasOutputFlagWithoutValue bool, isBatchMode bool) model.OutputConfig {
	config := model.OutputConfig{
		OutputDir: outputDir,
		Template:  outputTemplate,
		CreateDir: true,
	}
	
	// Handle special case where -o is used without arguments
	if hasOutputFlagWithoutValue || outputDir == "__BASENAME_SUBTITLES__" {
		if isBatchMode {
			config.OutputDir = "BATCH_BASENAME_SUBTITLES" // Special marker for batch mode
		} else {
			// For single file mode, we'll set this later when we know the input filename
			config.OutputDir = "__BASENAME_SUBTITLES__"
		}
	}
	
	if config.Template == "" {
		config.Template = model.DefaultOutputTemplate
	}
	
	return config
}

// ResolveOutputDirectory resolves special output directory markers based on the input file
func ResolveOutputDirectory(outputDir, inputFileName string) string {
	if outputDir == "__BASENAME_SUBTITLES__" || outputDir == "BATCH_BASENAME_SUBTITLES" {
		baseName := TrimExtension(filepath.Base(inputFileName))
		return filepath.Join(filepath.Dir(inputFileName), baseName+"-subtitles")
	}
	return outputDir
}

// TrimExtension removes the file extension from a filename
func TrimExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}