package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Processor handles CSV file processing operations.
type Processor struct {
	cleaner *Cleaner
}

// NewProcessor creates a new CSV processor instance.
func NewProcessor() *Processor {
	return &Processor{
		cleaner: NewCleaner(),
	}
}

// ProcessFile reads a CSV file, cleans it, and writes the result to an output file.
func (p *Processor) ProcessFile(inputPath, outputPath string) error {
	if err := p.validatePaths(inputPath, outputPath); err != nil {
		return err
	}
	
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input file %s: %w", inputPath, err)
	}
	defer inFile.Close()
	
	rawData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file %s: %w", inputPath, err)
	}
	
	cleanedData := p.cleaner.CleanData(string(rawData))
	
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %s: %w", outputPath, err)
	}
	defer outFile.Close()
	
	reader := csv.NewReader(strings.NewReader(cleanedData))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	
	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	
	if err := p.cleaner.CleanRecords(reader, writer); err != nil {
		return fmt.Errorf("clean records: %w", err)
	}
	
	return nil
}

func (p *Processor) validatePaths(inputPath, outputPath string) error {
	if inputPath == "" {
		return fmt.Errorf("input path cannot be empty")
	}
	
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}
	
	return nil
}
