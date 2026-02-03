package models

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"golang-csv-handler/helpers"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// FileInfo represents information about a file in S3.
// This is a simple struct to hold file metadata.
type FileInfo struct {
	Name         string // File name/path in S3
	Size         int64  // File size in bytes
	LastModified string // Last modified date/time as a formatted string
}

// LambdaResponse represents the response that Lambda will return.
// This structure will be automatically converted to JSON by Lambda.
type LambdaResponse struct {
	StatusCode int        `json:"statusCode"` // HTTP-like status code (200 = success)
	Message    string     `json:"message"`    // Human-readable message
	FileCount  int        `json:"fileCount"`  // Number of files found
	Files      []FileInfo `json:"files"`      // List of files found
}

// S3Checker handles checking for files in an S3 bucket.
// This struct holds the S3 client and bucket name.
type S3Checker struct {
	S3Client *s3.Client // S3 client used to make API calls to S3
	Bucket   string     // Name of the S3 bucket to check
}

// ProcessFile processes a single CSV file from S3.
// It reads the file, cleans it, and uploads the cleaned version.
func (s *S3Checker) ProcessFile(ctx context.Context, file FileInfo) (FileInfo, error) {
	log.Printf("Processing file: %s", file.Name)
	log.Printf("Bucket: %s", s.Bucket)
	log.Printf("Key: %s", file.Name)

	// Get the object from S3
	result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
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
	cleanedCSV := helpers.CleanCSVData(dirtyCSV)

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

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String("output/" + strings.Split(file.Name, "/")[1] + "_cleaned.csv"),
		Body:   bytes.NewReader([]byte(cleanedCSV)),
	})
	if err != nil {
		return file, fmt.Errorf("failed to put object: %w", err)
	}

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String("archive/" + strings.Split(file.Name, "/")[1]),
		Body:   bytes.NewReader([]byte(file.Name)),
	})
	if err != nil {
		return file, fmt.Errorf("failed to put object: %w", err)
	}

	log.Printf("Object put inside Output folder successfully")

	_, err = s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(file.Name),
	})
	if err != nil {
		return file, fmt.Errorf("failed to delete object: %w", err)
	}

	log.Printf("Object deleted from input folder successfully")

	return file, nil
}

// CheckForFiles checks for files in the S3 bucket.
// It lists all objects in the input/ prefix and returns their metadata.
func (s *S3Checker) CheckForFiles(ctx context.Context) ([]FileInfo, error) {
	// ListObjectsV2Input is the input structure for listing objects in S3.
	// We specify the bucket name we want to list objects from.
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket), // The bucket to list objects from
		Prefix: aws.String("input/"), // The prefix to filter by
		// MaxKeys: You can limit the number of results (optional)
		// Prefix: You can filter by prefix (e.g., "input/") (optional)
	}

	// Call the ListObjectsV2 API to get objects from the bucket.
	// This is an API call to AWS S3 service.
	result, err := s.S3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", s.Bucket, err)
	}

	// Convert S3 objects to our FileInfo structure.
	var files []FileInfo
	for _, obj := range result.Contents {
		// Skip folder markers (objects that end with '/' or have 0 size)
		if obj.Key != nil && (strings.HasSuffix(*obj.Key, "/") || (obj.Size != nil && *obj.Size == 0)) {
			continue
		}
		// obj is of type types.Object, which contains:
		// - Key: The file name/path in S3 (pointer to string)
		// - Size: File size in bytes (pointer to int64)
		// - LastModified: When the file was last modified
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		files = append(files, FileInfo{
			Name:         *obj.Key,                                       // File name/path
			Size:         size,                                           // File size in bytes (dereferenced from pointer)
			LastModified: obj.LastModified.Format("2006-01-02 15:04:05"), // Format timestamp
		})
	}

	return files, nil
}

// NewS3Checker creates a new S3Checker instance.
//
// Parameters:
//   - bucketName: The name of the S3 bucket to check for files
//
// Returns: *S3Checker and an error
//
// This function:
// 1. Loads AWS configuration (credentials, region) automatically
// 2. Creates an S3 client to interact with S3
// 3. Returns a configured S3Checker ready to use
func NewS3Checker(bucketName string) (*S3Checker, error) {
	// Load AWS configuration. This automatically:
	// - Reads AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from environment (if set)
	// - Reads from ~/.aws/credentials file (if exists)
	// - Uses IAM role credentials (if running in Lambda/EC2)
	// - Reads AWS_REGION from environment or config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an S3 client using the loaded configuration.
	// The client is used to make API calls to S3 (list objects, get objects, etc.)
	s3Client := s3.NewFromConfig(cfg)

	return &S3Checker{
		S3Client: s3Client,
		Bucket:   bucketName,
	}, nil
}
