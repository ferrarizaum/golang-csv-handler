package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type CSVHandler struct {
	inputFile  string
	outputFile string
}

func NewCSVHandler(inputFile, outputFile string) *CSVHandler {
	return &CSVHandler{
		inputFile:  inputFile,
		outputFile: outputFile,
	}
}

func (h *CSVHandler) CleanCSV() error {
	inFile, err := os.Open(h.inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(h.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	reader := csv.NewReader(inFile)

	writer := csv.NewWriter(outFile)

	defer writer.Flush()

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	cleanedHeader := h.cleanRow(header)

	if err := writer.Write(cleanedHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	seenRows := make(map[string]bool)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		cleanedRow := h.cleanRow(record)

		if h.isEmptyRow(cleanedRow) {
			continue
		}

		rowKey := strings.Join(cleanedRow, "|")

		if seenRows[rowKey] {
			continue
		}

		seenRows[rowKey] = true

		if err := writer.Write(cleanedRow); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (h *CSVHandler) cleanRow(row []string) []string {
	cleaned := make([]string, len(row))

	for i, field := range row {
		cleaned[i] = strings.TrimSpace(field)

		cleaned[i] = h.removeTrashChars(cleaned[i])
	}
	return cleaned
}

func (h *CSVHandler) removeTrashChars(s string) string {
	var result strings.Builder

	for _, r := range s {
		if r >= 32 && r <= 126 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (h *CSVHandler) isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func main() {
	inputFile := flag.String("input", "", "Path to input CSV file (required)")
	outputFile := flag.String("output", "", "Path to output CSV file (required)")

	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: golang-csv-handler -input <input.csv> -output <output.csv>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("Error: Input file '%s' does not exist\n", *inputFile)
		os.Exit(1)
	}

	handler := NewCSVHandler(*inputFile, *outputFile)

	fmt.Printf("Processing CSV file: %s\n", *inputFile)

	if err := handler.CleanCSV(); err != nil {
		fmt.Printf("Error processing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully cleaned CSV saved to: %s\n", *outputFile)
}
