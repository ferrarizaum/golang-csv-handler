package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"golang-csv-handler/internal/s3"
)

const (
	envBucketName = "S3_BUCKET_NAME"
	statusOK      = 200
	statusError   = 500
)

// Event represents the Lambda event payload (currently unused but kept for extensibility).
type Event struct{}

// Response represents the Lambda function response.
type Response struct {
	StatusCode int           `json:"statusCode"`
	Message    string        `json:"message"`
	FileCount  int           `json:"fileCount,omitempty"`
	Files      []s3.FileInfo `json:"files,omitempty"`
}

type handler struct {
	s3Client *s3.Client
}

func main() {
	h, err := newHandler()
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}
	
	lambda.Start(h.handle)
}

func newHandler() (*handler, error) {
	bucketName := os.Getenv(envBucketName)
	if bucketName == "" {
		return nil, fmt.Errorf("environment variable %s is not set", envBucketName)
	}
	
	ctx := context.Background()
	client, err := s3.NewClient(ctx, s3.Config{
		Bucket: bucketName,
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}
	
	return &handler{
		s3Client: client,
	}, nil
}

func (h *handler) handle(ctx context.Context, event Event) (Response, error) {
	log.Println("Lambda function invoked")
	
	files, err := h.s3Client.ListFiles(ctx)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to list files: %v", err)), err
	}
	
	log.Printf("Found %d file(s)", len(files))
	
	if len(files) == 0 {
		return h.successResponse("No files found", files), nil
	}
	
	for _, file := range files {
		log.Printf("Processing: %s (Size: %d bytes)", file.Name, file.Size)
		
		if err := h.s3Client.ProcessFile(ctx, file); err != nil {
			log.Printf("Failed to process file %s: %v", file.Name, err)
			continue
		}
		
		log.Printf("Successfully processed: %s", file.Name)
	}
	
	return h.successResponse(
		fmt.Sprintf("Successfully processed %d file(s)", len(files)),
		files,
	), nil
}

func (h *handler) successResponse(message string, files []s3.FileInfo) Response {
	return Response{
		StatusCode: statusOK,
		Message:    message,
		FileCount:  len(files),
		Files:      files,
	}
}

func (h *handler) errorResponse(message string) Response {
	return Response{
		StatusCode: statusError,
		Message:    message,
	}
}
