// Package main contains the AWS Lambda handler function.
//
// AWS Lambda is a serverless compute service that runs your code in response to events.
// This Lambda function is triggered by EventBridge (a scheduler) to check for files in S3.
//
// How Lambda works:
// 1. You upload your code to AWS Lambda
// 2. AWS runs your code when triggered (by EventBridge in our case)
// 3. Lambda automatically manages the servers, scaling, and infrastructure
// 4. You only pay for the compute time you use
package main

import (
	// context: Provides request-scoped values, cancellation signals, and deadlines.
	// In Lambda, the context contains information about the invocation request
	// and allows you to handle timeouts and cancellations.

	"context"

	// fmt: Formatting package for printing and formatting strings.
	"fmt"

	// log: Simple logging package. In Lambda, logs go to CloudWatch Logs.
	// CloudWatch is AWS's monitoring and logging service.
	"log"

	// os: Operating system interface. We use it to read environment variables.
	// Environment variables are a way to pass configuration to Lambda functions
	// without hardcoding values in your code.
	"os"

	// github.com/aws/aws-lambda-go/lambda: AWS Lambda Go runtime library.
	// This package provides the Lambda handler interface and runtime.
	// It handles the communication between AWS Lambda and your Go code.
	"github.com/aws/aws-lambda-go/lambda"

	// github.com/aws/aws-sdk-go-v2/aws: AWS SDK configuration and utilities.
	// The AWS SDK is a library that provides APIs to interact with AWS services.

	// golang-csv-handler/models: Our models package containing S3Checker and data structures
	"golang-csv-handler/models"
)

// LambdaRequest represents the event that triggers the Lambda.
// EventBridge scheduler sends a JSON event, but for a simple scheduler trigger,
// we might not need any specific fields. This struct allows for future expansion.
type LambdaRequest struct {
	// EventBridge scheduler events are typically simple, but we can add fields here
	// if we need to pass information from the scheduler to the Lambda.
}

// Handler is the main Lambda handler function.
// This is the entry point that AWS Lambda calls when the function is invoked.
//
// Parameters:
//   - ctx: Context provides request-scoped values and cancellation signals
//   - event: The event that triggered the Lambda (from EventBridge in our case)
//
// Returns: LambdaResponse and an error
//
// How it works:
// 1. EventBridge scheduler triggers this function at the scheduled time
// 2. We read the S3 bucket name from an environment variable
// 3. We check for files in the S3 bucket
// 4. We return a response with the results
func Handler(ctx context.Context, event LambdaRequest) (models.LambdaResponse, error) {
	// Log that the function was invoked. These logs appear in CloudWatch Logs.
	log.Printf("Lambda function invoked. Checking S3 bucket for files...")

	// Get configuration from environment variables.
	// Environment variables are set in the Lambda function configuration.
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    "S3_BUCKET_NAME environment variable is not set",
		}, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	log.Printf("Checking bucket: %s", bucketName)

	// Create an S3Checker instance to interact with S3.
	checker, err := models.NewS3Checker(bucketName)
	if err != nil {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create S3 checker: %v", err),
		}, err
	}

	// Check for files in the S3 bucket.
	files, err := checker.CheckForFiles(ctx)
	if err != nil {
		return models.LambdaResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to check S3 bucket: %v", err),
		}, err
	}

	// Log the results. This helps with debugging and monitoring.
	log.Printf("Found %d file(s) in bucket %s", len(files), bucketName)

	// If files were found, trigger ECS tasks to process them
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

	// Return a successful response with the file information.
	// Lambda will automatically convert this struct to JSON.
	return models.LambdaResponse{
		StatusCode: 200,
		Message:    fmt.Sprintf("Successfully checked bucket. Found %d file(s)", len(files)),
		FileCount:  len(files),
		Files:      files,
	}, nil
}

// main is the entry point when running locally (for testing).
// In Lambda, AWS calls the Handler function directly, but main() is still needed
// to register the handler with the Lambda runtime.
func main() {
	// lambda.Start() is the entry point for AWS Lambda.
	// It takes your handler function and starts the Lambda runtime.
	// The runtime handles:
	// - Receiving events from AWS
	// - Calling your handler function
	// - Returning the response
	// - Managing the execution environment
	lambda.Start(Handler)
}
