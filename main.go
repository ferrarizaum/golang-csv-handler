// Package main is the entry point of a Go program.
// When you run "go run" or "go build", Go looks for the main package
// and executes the main() function inside it.
package main

import (
	// encoding/csv: Provides functionality to read and write CSV (Comma-Separated Values) files.
	// It handles parsing CSV format, escaping quotes, and managing delimiters automatically.
	"encoding/csv"

	// flag: Package for parsing command-line flags/arguments.
	// Allows you to define flags like -input and -output that users can pass when running the program.
	"flag"

	// fmt: Formatting package for printing to console and formatting strings.
	// Functions like Printf, Println, and Errorf come from this package.
	"fmt"

	// io: Provides basic I/O (Input/Output) primitives.
	// We use io.EOF (End Of File) constant to detect when we've finished reading a file.
	"io"

	// os: Operating system interface. Provides functions to interact with the OS.
	// Used for opening files (os.Open), creating files (os.Create), checking file existence (os.Stat), etc.
	"os"

	// strings: String manipulation functions.
	// Provides TrimSpace (remove whitespace), Join (combine strings), and other string utilities.
	"strings"
)

// CSVHandler is a struct (like a class in other languages) that holds data and methods
// for handling CSV file operations. A struct groups related data together.
//
// In Go, structs are defined with the "type" keyword followed by the name and "struct".
// Fields are defined inside curly braces.
type CSVHandler struct {
	inputFile  string // Path to the input CSV file that needs to be cleaned
	outputFile string // Path where the cleaned CSV file will be saved
}

// NewCSVHandler is a constructor function that creates and returns a new CSVHandler instance.
//
// Parameters:
//   - inputFile: The path to the CSV file to be cleaned
//   - outputFile: The path where the cleaned CSV will be saved
//
// Returns: *CSVHandler (a pointer to CSVHandler)
//
// Note: The asterisk (*) means this function returns a pointer.
// Pointers in Go allow you to pass references to data instead of copying it,
// which is more efficient for larger data structures.
func NewCSVHandler(inputFile, outputFile string) *CSVHandler {
	// The & operator creates a pointer to the CSVHandler struct.
	// This returns the memory address of the struct, not a copy of it.
	return &CSVHandler{
		inputFile:  inputFile,
		outputFile: outputFile,
	}
}

// CleanCSV is a method (function attached to a struct) that processes the CSV file
// and removes unwanted data. The (h *CSVHandler) part means this function belongs to CSVHandler.
//
// The asterisk (*) before CSVHandler means we're using a pointer receiver,
// which allows the method to modify the struct if needed (and is more efficient).
//
// Returns: error - In Go, functions that can fail return an error as the last return value.
// If everything succeeds, we return nil (which means "no error").
func (h *CSVHandler) CleanCSV() error {
	// os.Open opens a file for reading. It returns a file handle and an error.
	// In Go, errors are values, not exceptions, so we check them explicitly.
	inFile, err := os.Open(h.inputFile)
	if err != nil {
		// fmt.Errorf creates a formatted error message.
		// The %w verb wraps the original error, preserving it for error inspection.
		return fmt.Errorf("failed to open input file: %w", err)
	}
	// defer schedules a function call to be executed when the surrounding function returns.
	// This ensures the file is always closed, even if an error occurs later.
	// It's Go's way of handling cleanup (like "finally" in other languages).
	defer inFile.Close()

	// os.Create creates a new file for writing. If the file exists, it truncates it.
	outFile, err := os.Create(h.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// csv.NewReader creates a CSV reader that reads from the input file.
	// It automatically handles CSV parsing, quotes, and escaping.
	reader := csv.NewReader(inFile)

	// csv.NewWriter creates a CSV writer that writes to the output file.
	writer := csv.NewWriter(outFile)

	// Flush writes any buffered data to the file.
	// We defer it to ensure all data is written even if the function returns early.
	defer writer.Flush()

	// Read the first row, which should be the header (column names).
	// reader.Read() returns a slice of strings ([]string) representing the row fields.
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Clean the header row (trim whitespace, remove trash characters).
	cleanedHeader := h.cleanRow(header)

	// Write the cleaned header to the output file.
	if err := writer.Write(cleanedHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create a map to track rows we've already seen.
	// A map is like a dictionary: it stores key-value pairs.
	// map[string]bool means: keys are strings, values are booleans.
	// We'll use this to detect and skip duplicate rows.
	seenRows := make(map[string]bool)

	// Infinite loop that continues until we break out of it.
	// This is Go's way of reading a file until the end.
	for {
		// Read the next row from the CSV file.
		record, err := reader.Read()

		// io.EOF (End Of File) is a special error that means we've reached the end of the file.
		// This is not a real error - it's expected when we finish reading.
		if err == io.EOF {
			break // Exit the loop when we reach the end of the file
		}
		// If there's any other error, return it.
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		// Clean the row: trim whitespace and remove unwanted characters.
		cleanedRow := h.cleanRow(record)

		// Skip rows where all fields are empty (no useful data).
		if h.isEmptyRow(cleanedRow) {
			continue // Skip to the next iteration of the loop
		}

		// Create a unique key for this row by joining all fields with "|".
		// This allows us to detect duplicate rows.
		// Example: ["John", "john@email.com", "30"] becomes "John|john@email.com|30"
		rowKey := strings.Join(cleanedRow, "|")

		// Check if we've seen this exact row before.
		// If the key exists in the map, the row is a duplicate.
		if seenRows[rowKey] {
			continue // Skip duplicate rows
		}

		// Mark this row as seen by storing it in the map.
		seenRows[rowKey] = true

		// Write the cleaned row to the output CSV file.
		if err := writer.Write(cleanedRow); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	// Return nil to indicate success (no error occurred).
	return nil
}

// cleanRow processes each field in a CSV row and removes unwanted data.
//
// Parameters:
//   - row: A slice (like an array) of strings representing one row of CSV data
//
// Returns: A new slice with cleaned strings
//
// What it does:
//  1. Trims whitespace from each field
//  2. Removes control characters and unwanted symbols
func (h *CSVHandler) cleanRow(row []string) []string {
	// make() creates a new slice with the specified length.
	// len(row) gets the number of elements in the row slice.
	cleaned := make([]string, len(row))

	// Range loop: iterates over each element in the slice.
	// i is the index (0, 1, 2, ...), field is the value at that index.
	// This is Go's equivalent of "for each" loops in other languages.
	for i, field := range row {
		// strings.TrimSpace removes leading and trailing whitespace (spaces, tabs, newlines).
		// Example: "  hello  " becomes "hello"
		cleaned[i] = strings.TrimSpace(field)

		// Remove unwanted characters (control characters, special symbols, etc.)
		cleaned[i] = h.removeTrashChars(cleaned[i])
	}
	return cleaned
}

// removeTrashChars filters out unwanted characters from a string field.
//
// Parameters:
//   - s: The string to clean
//
// Returns: A new string with only allowed characters
//
// What characters are kept:
//   - Printable ASCII characters (codes 32-126): letters, numbers, punctuation
//   - Tab (\t), newline (\n), and carriage return (\r) characters
//
// What characters are removed:
//   - Control characters (codes 0-31 except tab/newline/carriage return)
//   - Non-printable characters
func (h *CSVHandler) removeTrashChars(s string) string {
	// strings.Builder is an efficient way to build strings in Go.
	// It's better than concatenating strings with "+" because it doesn't
	// create new strings each time (which is inefficient).
	var result strings.Builder

	// Range over a string iterates over Unicode code points (runes), not bytes.
	// In Go, a string is a sequence of bytes, but when you range over it,
	// you get runes (Unicode characters).
	for _, r := range s {
		// Check if the character is allowed:
		// - ASCII printable range: 32 (space) to 126 (~)
		// - OR it's a tab (\t), newline (\n), or carriage return (\r)
		// The || operator means "OR" in Go
		if r >= 32 && r <= 126 || r == '\t' || r == '\n' || r == '\r' {
			// WriteRune adds the character to the builder if it's allowed.
			result.WriteRune(r)
		}
		// Characters that don't meet the condition are simply not added (filtered out).
		// You can add more specific rules here based on your needs.
	}

	// String() converts the builder to a final string.
	return result.String()
}

// isEmptyRow checks if a CSV row contains only empty fields.
//
// Parameters:
//   - row: A slice of strings representing one row of CSV data
//
// Returns: true if all fields are empty (after trimming whitespace), false otherwise
//
// This function is used to filter out rows that have no useful data.
func (h *CSVHandler) isEmptyRow(row []string) bool {
	// Iterate through each field in the row.
	for _, field := range row {
		// Trim whitespace and check if the field is not empty.
		// If we find even one non-empty field, the row is not empty.
		if strings.TrimSpace(field) != "" {
			return false // Found a non-empty field, so row is not empty
		}
	}
	// If we get here, all fields were empty.
	return true // All fields are empty
}

// main is the entry point of every Go program.
// When you run the program, Go automatically calls this function.
func main() {
	// flag.String defines a command-line flag that accepts a string value.
	// Parameters:
	//   - "input": the flag name (users will use -input)
	//   - "": the default value (empty string means no default)
	//   - "Path to input CSV file (required)": help text shown when user uses -help
	//
	// flag.String returns a pointer to a string (*string).
	// The asterisk (*) means it's a pointer - a reference to the actual value.
	inputFile := flag.String("input", "", "Path to input CSV file (required)")
	outputFile := flag.String("output", "", "Path to output CSV file (required)")

	// flag.Parse() reads the command-line arguments and fills in the flag values.
	// After calling this, *inputFile and *outputFile will contain the values
	// the user provided (or empty strings if not provided).
	flag.Parse()

	// Validate that both required flags were provided.
	// The * operator dereferences the pointer to get the actual value.
	// If either is empty, we show usage instructions and exit.
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: golang-csv-handler -input <input.csv> -output <output.csv>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults() // Prints all flag definitions with their help text
		os.Exit(1)           // Exit with error code 1 (non-zero means error)
	}

	// os.Stat gets information about a file (like checking if it exists).
	// It returns file info and an error. We use _ to ignore the file info.
	// os.IsNotExist checks if the error means the file doesn't exist.
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("Error: Input file '%s' does not exist\n", *inputFile)
		os.Exit(1)
	}

	// Create a new CSVHandler instance using our constructor function.
	// We dereference the pointers (*inputFile, *outputFile) to get the actual string values.
	handler := NewCSVHandler(*inputFile, *outputFile)

	// Print a message to let the user know processing has started.
	fmt.Printf("Processing CSV file: %s\n", *inputFile)

	// Call the CleanCSV method to process the file.
	// If it returns an error, we print it and exit with an error code.
	if err := handler.CleanCSV(); err != nil {
		// %v is a verb that formats the error in a default way.
		fmt.Printf("Error processing CSV: %v\n", err)
		os.Exit(1)
	}

	// If we get here, everything succeeded!
	fmt.Printf("Successfully cleaned CSV saved to: %s\n", *outputFile)
}
