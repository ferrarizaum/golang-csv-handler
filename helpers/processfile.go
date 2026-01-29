package helpers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Checker) ProcessFile(ctx context.Context, file FileInfo) (FileInfo, error) {
	log.Printf("Processing file: %s", file.Name)
	log.Printf("Bucket: %s", s.bucket)
	log.Printf("Key: %s", file.Name)

	// Get the object from S3
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.Name),
	})
	if err != nil {
		return file, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	// Read the entire file content as raw bytes
	rawData, err := io.ReadAll(result.Body)
	if err != nil {
		return file, fmt.Errorf("failed to read file content: %w", err)
	}

	log.Printf("Original file size: %d bytes", len(rawData))

	// Convert to string and clean the data
	dirtyCSV := string(rawData)
	cleanedCSV := cleanCSVData(dirtyCSV)

	log.Printf("Cleaned file size: %d bytes", len(cleanedCSV))

	// Create a CSV reader from the cleaned data
	csvReader := csv.NewReader(bytes.NewReader([]byte(cleanedCSV)))

	// Configure the CSV reader to be more lenient with any remaining issues
	csvReader.LazyQuotes = true       // Allow bare quotes in unquoted fields
	csvReader.TrimLeadingSpace = true // Trim leading space in fields
	csvReader.FieldsPerRecord = -1    // Allow variable number of fields per record

	// Read all records from the cleaned CSV
	records, err := csvReader.ReadAll()
	if err != nil {
		return file, fmt.Errorf("failed to parse cleaned CSV: %w", err)
	}

	// Log the CSV content
	log.Printf("CSV file has %d rows after cleaning", len(records))
	for i, record := range records {
		log.Printf("Row %d (%d fields): %v", i, len(record), record)
	}

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String("output/" + strings.Split(file.Name, "/")[1] + "_cleaned.csv"),
		Body:   bytes.NewReader([]byte(cleanedCSV)),
	})
	if err != nil {
		return file, fmt.Errorf("failed to put object: %w", err)
	}

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String("archive/" + strings.Split(file.Name, "/")[1]),
		Body:   bytes.NewReader([]byte(file.Name)),
	})
	if err != nil {
		return file, fmt.Errorf("failed to put object: %w", err)
	}

	log.Printf("Object put inside Output folder successfully")

	_, err = s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.Name),
	})
	if err != nil {
		return file, fmt.Errorf("failed to delete object: %w", err)
	}

	log.Printf("Object deleted from input folder successfully")

	return file, nil
}
