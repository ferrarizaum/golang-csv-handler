package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"golang-csv-handler/models"
)

type LambdaRequest struct {
}

func Handler(ctx context.Context, event LambdaRequest) (models.LambdaResponse, error) {
	log.Printf("Lambda function invoked. Checking S3 bucket for files...")

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    "S3_BUCKET_NAME environment variable is not set",
		}, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	log.Printf("Checking bucket: %s", bucketName)

	checker, err := models.NewS3Checker(bucketName)
	if err != nil {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create S3 checker: %v", err),
		}, err
	}

	files, err := checker.CheckForFiles(ctx)
	if err != nil {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to check S3 bucket: %v", err),
		}, err
	}

	log.Printf("Found %d file(s) in bucket %s", len(files), bucketName)

	if len(files) > 0 {
		log.Println("Files found:")
		for _, file := range files {
			log.Printf("  - %s (Size: %d bytes, Modified: %s)", file.Name, file.Size, file.LastModified)
			processedFile, err := checker.ProcessFile(ctx, file)
			if err != nil {
				log.Printf("Failed to process file: %v", err)
			}
			log.Printf("Processed file: %s", processedFile.Name)
		}

	} else {
		log.Println("No files found in the bucket.")
	}

	return models.LambdaResponse{
		StatusCode: 200,
		Message:    fmt.Sprintf("Successfully checked bucket. Found %d file(s)", len(files)),
		FileCount:  len(files),
		Files:      files,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
