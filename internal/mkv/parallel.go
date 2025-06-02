package mkv

import (
	"fmt"
	"sync"

	"subscalpelmkv/internal/format"
	"subscalpelmkv/internal/model"
)

// ExtractionJob represents a single subtitle extraction task
type ExtractionJob struct {
	Track         model.MKVTrack
	OriginalTrack model.MKVTrack
	OutFileName   string
	MksFileName   string
}

// ExtractionResult represents the result of an extraction operation
type ExtractionResult struct {
	Job   ExtractionJob
	Error error
}

// ExtractSubtitlesParallel extracts multiple subtitle tracks concurrently
// This is more efficient than sequential extraction when dealing with multiple tracks
func ExtractSubtitlesParallel(jobs []ExtractionJob, maxWorkers int) []ExtractionResult {
	if len(jobs) == 0 {
		return []ExtractionResult{}
	}

	// For single track, no need for parallelization overhead
	if len(jobs) == 1 {
		job := jobs[0]
		err := ExtractSubtitles(job.MksFileName, job.Track, job.OutFileName, job.OriginalTrack.Properties.Number)
		return []ExtractionResult{{Job: job, Error: err}}
	}

	// Create channels for job distribution and result collection
	jobChan := make(chan ExtractionJob, len(jobs))
	resultChan := make(chan ExtractionResult, len(jobs))

	// Limit concurrent workers to prevent overwhelming the system
	if maxWorkers <= 0 || maxWorkers > len(jobs) {
		maxWorkers = calculateOptimalWorkers(len(jobs))
	}

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Process jobs from the channel
			for job := range jobChan {
				// Perform the actual extraction
				err := ExtractSubtitles(job.MksFileName, job.Track, job.OutFileName, job.OriginalTrack.Properties.Number)

				// Send result back
				resultChan <- ExtractionResult{
					Job:   job,
					Error: err,
				}
			}
		}(i)
	}

	// Send all jobs to workers
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []ExtractionResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// calculateOptimalWorkers determines the best number of concurrent workers
func calculateOptimalWorkers(trackCount int) int {
	// Conservative approach: limit concurrent mkvextract processes
	// Each mkvextract process can be I/O intensive

	if trackCount == 1 {
		return 1 // No benefit from parallelization
	}

	if trackCount <= 4 {
		return trackCount // One worker per track for small counts
	}

	// For larger track counts, limit to 4 concurrent processes
	// This prevents overwhelming the disk I/O and system resources
	// You can adjust this based on your system's capabilities
	return 4
}

// ExtractSubtitlesParallelWithProgress is a wrapper that provides progress feedback
func ExtractSubtitlesParallelWithProgress(jobs []ExtractionJob, maxWorkers int) error {
	if len(jobs) == 0 {
		format.PrintWarning("No subtitle tracks to extract")
		return nil
	}

	results := ExtractSubtitlesParallel(jobs, maxWorkers)

	// Process results and check for errors
	successCount := 0
	for _, result := range results {
		if result.Error != nil {
			format.PrintError(fmt.Sprintf("Error extracting track %d: %v",
				result.Job.OriginalTrack.Properties.Number, result.Error))
			return result.Error
		}
		successCount++
	}

	fmt.Println()
	if successCount == 0 {
		format.PrintWarning("No subtitle tracks were extracted")
	} else {
		format.PrintSuccess(fmt.Sprintf("Successfully extracted %d subtitle track(s)",
			successCount))
	}

	return nil
}
