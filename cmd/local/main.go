package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang-csv-handler/internal/csv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

type fileJob struct {
	inputPath  string
	outputPath string
}

type fileResult struct {
	filename string
	err      error
}

func run() error {
	var (
		inputFolder  = flag.String("input", "", "Path to input CSV folder (required)")
		outputFolder = flag.String("output", "", "Path to output CSV folder (required)")
		workers      = flag.Int("workers", 4, "Number of concurrent workers")
	)

	flag.Parse()

	if *inputFolder == "" || *outputFolder == "" {
		flag.Usage()
		return fmt.Errorf("both -input and -output flags are required")
	}

	if err := os.MkdirAll(*outputFolder, 0755); err != nil {
		return fmt.Errorf("create output folder: %w", err)
	}

	files, err := os.ReadDir(*inputFolder)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	csvFiles := filterCSVFiles(files)
	if len(csvFiles) == 0 {
		log.Println("No CSV files found in input folder")
		return nil
	}

	log.Printf("Found %d CSV file(s) to process using %d worker(s)", len(csvFiles), *workers)

	jobs := make(chan fileJob, len(csvFiles))
	results := make(chan fileResult, len(csvFiles))

	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(i+1, jobs, results, &wg)
	}

	for _, file := range csvFiles {
		inputPath := filepath.Join(*inputFolder, file.Name())
		outputPath := filepath.Join(*outputFolder, file.Name())
		jobs <- fileJob{
			inputPath:  inputPath,
			outputPath: outputPath,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	errorCount := 0
	for result := range results {
		if result.err != nil {
			log.Printf("❌ Failed to process %s: %v", result.filename, result.err)
			errorCount++
		} else {
			log.Printf("✓ Successfully processed: %s", result.filename)
			successCount++
		}
	}

	log.Printf("\nProcessing complete: %d succeeded, %d failed", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) failed to process", errorCount)
	}

	return nil
}

func worker(id int, jobs <-chan fileJob, results chan<- fileResult, wg *sync.WaitGroup) {
	defer wg.Done()

	processor := csv.NewProcessor()

	for job := range jobs {
		log.Printf("Worker %d: Processing %s", id, filepath.Base(job.inputPath))

		err := processor.ProcessFile(job.inputPath, job.outputPath)

		results <- fileResult{
			filename: filepath.Base(job.inputPath),
			err:      err,
		}
	}
}

func filterCSVFiles(files []os.DirEntry) []os.DirEntry {
	csvFiles := make([]os.DirEntry, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".csv" {
			csvFiles = append(csvFiles, file)
		}
	}
	return csvFiles
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintf(os.Stderr, "  %s -input ./input -output ./output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input ./input -output ./output -workers 8\n", os.Args[0])
	}
}
