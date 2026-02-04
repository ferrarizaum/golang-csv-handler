package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang-csv-handler/internal/csv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	var (
		inputFile  = flag.String("input", "", "Path to input CSV file (required)")
		outputFile = flag.String("output", "", "Path to output CSV file (required)")
	)
	
	flag.Parse()
	
	if *inputFile == "" || *outputFile == "" {
		flag.Usage()
		return fmt.Errorf("both -input and -output flags are required")
	}
	
	processor := csv.NewProcessor()
	
	log.Printf("Processing CSV file: %s", *inputFile)
	
	if err := processor.ProcessFile(*inputFile, *outputFile); err != nil {
		return fmt.Errorf("process file: %w", err)
	}
	
	log.Printf("Successfully cleaned CSV saved to: %s", *outputFile)
	
	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExample:")
		fmt.Fprintf(os.Stderr, "  %s -input data.csv -output clean.csv\n", os.Args[0])
	}
}
