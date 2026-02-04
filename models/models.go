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

type FileInfo struct {
	Name         string
	Size         int64
	LastModified string
}

type LambdaResponse struct {
	StatusCode int        `json:"statusCode"`
	Message    string     `json:"message"`
	FileCount  int        `json:"fileCount"`
	Files      []FileInfo `json:"files"`
}

type S3Checker struct {
	S3Client *s3.Client
	Bucket   string
}

func (s *S3Checker) ProcessFile(ctx context.Context, file FileInfo) (FileInfo, error) {
	log.Printf("Processing file: %s", file.Name)
	log.Printf("Bucket: %s", s.Bucket)
	log.Printf("Key: %s", file.Name)

	result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(file.Name),
	})
	if err != nil {
		return file, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	rawData, err := io.ReadAll(result.Body)
	if err != nil {
		return file, fmt.Errorf("failed to read file content: %w", err)
	}

	log.Printf("Original file size: %d bytes", len(rawData))

	dirtyCSV := string(rawData)
	cleanedCSV := helpers.CleanCSVData(dirtyCSV)

	log.Printf("Cleaned file size: %d bytes", len(cleanedCSV))

	csvReader := csv.NewReader(bytes.NewReader([]byte(cleanedCSV)))

	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	records, err := csvReader.ReadAll()
	if err != nil {
		return file, fmt.Errorf("failed to parse cleaned CSV: %w", err)
	}

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

func (s *S3Checker) CheckForFiles(ctx context.Context) ([]FileInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String("input/"),
	}

	result, err := s.S3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", s.Bucket, err)
	}

	var files []FileInfo
	for _, obj := range result.Contents {
		if obj.Key != nil && (strings.HasSuffix(*obj.Key, "/") || (obj.Size != nil && *obj.Size == 0)) {
			continue
		}
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		files = append(files, FileInfo{
			Name:         *obj.Key,
			Size:         size,
			LastModified: obj.LastModified.Format("2006-01-02 15:04:05"),
		})
	}

	return files, nil
}

func NewS3Checker(bucketName string) (*S3Checker, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	return &S3Checker{
		S3Client: s3Client,
		Bucket:   bucketName,
	}, nil
}
