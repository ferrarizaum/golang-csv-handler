package s3

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	csvpkg "golang-csv-handler/internal/csv"
)

const (
	inputPrefix   = "input/"
	outputPrefix  = "output/"
	archivePrefix = "archive/"
	cleanedSuffix = "_cleaned.csv"
)

// FileInfo represents metadata about a file in S3.
type FileInfo struct {
	Name         string
	Size         int64
	LastModified time.Time
}

// Client handles S3 operations for CSV files.
type Client struct {
	s3Client *s3.Client
	bucket   string
	cleaner  *csvpkg.Cleaner
}

// Config holds configuration for the S3 client.
type Config struct {
	Bucket string
	Region string
}

// NewClient creates a new S3 client instance.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	
	return &Client{
		s3Client: s3.NewFromConfig(awsCfg),
		bucket:   cfg.Bucket,
		cleaner:  csvpkg.NewCleaner(),
	}, nil
}

// ListFiles lists all CSV files in the input folder.
func (c *Client) ListFiles(ctx context.Context) ([]FileInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(inputPrefix),
	}
	
	result, err := c.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list objects in bucket %s: %w", c.bucket, err)
	}
	
	files := make([]FileInfo, 0, len(result.Contents))
	
	for _, obj := range result.Contents {
		objTyped := types.Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		}
		
		if c.shouldSkipObject(&objTyped) {
			continue
		}
		
		files = append(files, FileInfo{
			Name:         aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}
	
	return files, nil
}

// ProcessFile downloads a file from S3, cleans it, uploads the cleaned version,
// archives the original, and deletes it from the input folder.
func (c *Client) ProcessFile(ctx context.Context, file FileInfo) error {
	log.Printf("Processing file: %s (size: %d bytes)", file.Name, file.Size)
	
	rawData, err := c.downloadFile(ctx, file.Name)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	
	cleanedData, err := c.cleanCSVData(rawData)
	if err != nil {
		return fmt.Errorf("clean CSV data: %w", err)
	}
	
	log.Printf("Cleaned file size: %d bytes (original: %d bytes)", len(cleanedData), len(rawData))
	
	outputKey := c.getOutputKey(file.Name)
	if err := c.uploadFile(ctx, outputKey, cleanedData); err != nil {
		return fmt.Errorf("upload cleaned file: %w", err)
	}
	
	log.Printf("Uploaded cleaned file to: %s", outputKey)
	
	archiveKey := c.getArchiveKey(file.Name)
	if err := c.uploadFile(ctx, archiveKey, rawData); err != nil {
		return fmt.Errorf("archive original file: %w", err)
	}
	
	log.Printf("Archived original file to: %s", archiveKey)
	
	if err := c.deleteFile(ctx, file.Name); err != nil {
		return fmt.Errorf("delete original file: %w", err)
	}
	
	log.Printf("Deleted original file from input folder: %s", file.Name)
	
	return nil
}

func (c *Client) downloadFile(ctx context.Context, key string) ([]byte, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer result.Body.Close()
	
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("read object body: %w", err)
	}
	
	return data, nil
}

func (c *Client) cleanCSVData(data []byte) ([]byte, error) {
	dirtyCSV := string(data)
	cleanedCSV := c.cleaner.CleanData(dirtyCSV)
	
	csvReader := csv.NewReader(bytes.NewReader([]byte(cleanedCSV)))
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1
	
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse cleaned CSV: %w", err)
	}
	
	log.Printf("CSV has %d rows after cleaning", len(records))
	
	return []byte(cleanedCSV), nil
}

func (c *Client) uploadFile(ctx context.Context, key string, data []byte) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("put object %s: %w", key, err)
	}
	
	return nil
}

func (c *Client) deleteFile(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %s: %w", key, err)
	}
	
	return nil
}

func (c *Client) shouldSkipObject(obj *types.Object) bool {
	if obj.Key == nil || obj.Size == nil {
		return true
	}
	
	if strings.HasSuffix(*obj.Key, "/") || *obj.Size == 0 {
		return true
	}
	
	return false
}

func (c *Client) getOutputKey(inputKey string) string {
	filename := filepath.Base(inputKey)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	return outputPrefix + nameWithoutExt + cleanedSuffix
}

func (c *Client) getArchiveKey(inputKey string) string {
	filename := filepath.Base(inputKey)
	return archivePrefix + filename
}
