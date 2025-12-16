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

	// strings: String manipulation functions.
	// Provides TrimSpace, Split, and other string utilities.
	"strings"

	// github.com/aws/aws-lambda-go/lambda: AWS Lambda Go runtime library.
	// This package provides the Lambda handler interface and runtime.
	// It handles the communication between AWS Lambda and your Go code.
	"github.com/aws/aws-lambda-go/lambda"

	// github.com/aws/aws-sdk-go-v2/aws: AWS SDK configuration and utilities.
	// The AWS SDK is a library that provides APIs to interact with AWS services.
	"github.com/aws/aws-sdk-go-v2/aws"

	// github.com/aws/aws-sdk-go-v2/config: Loads AWS configuration (credentials, region, etc.).
	// It automatically reads from environment variables, AWS credentials file, or IAM roles.
	"github.com/aws/aws-sdk-go-v2/config"

	// github.com/aws/aws-sdk-go-v2/service/s3: S3 service client.
	// S3 (Simple Storage Service) is AWS's object storage service (like a file system in the cloud).
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// github.com/aws/aws-sdk-go-v2/service/s3/types: S3-specific types and constants.

	// github.com/aws/aws-sdk-go-v2/service/ecs: ECS service client.
	// ECS (Elastic Container Service) is AWS's container orchestration service.
	// We use it to run Fargate tasks that process CSV files.
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// S3Checker handles checking for files in an S3 bucket.
// This struct holds the S3 client and bucket name.
type S3Checker struct {
	s3Client *s3.Client // S3 client used to make API calls to S3
	bucket   string     // Name of the S3 bucket to check
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
		s3Client: s3Client,
		bucket:   bucketName,
	}, nil
}

// CheckForFiles lists objects in the S3 bucket and returns information about them.
//
// Returns: A slice of file information and an error
//
// What this does:
// 1. Calls S3's ListObjectsV2 API to get a list of files in the bucket
// 2. Extracts file names, sizes, and last modified dates
// 3. Returns the information as a structured format
func (s *S3Checker) CheckForFiles(ctx context.Context) ([]FileInfo, error) {
	// ListObjectsV2Input is the input structure for listing objects in S3.
	// We specify the bucket name we want to list objects from.
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket), // The bucket to list objects from
		// MaxKeys: You can limit the number of results (optional)
		// Prefix: You can filter by prefix (e.g., "csv-files/") (optional)
	}

	// Call the ListObjectsV2 API to get objects from the bucket.
	// This is an API call to AWS S3 service.
	result, err := s.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", s.bucket, err)
	}

	// Convert S3 objects to our FileInfo structure.
	var files []FileInfo
	for _, obj := range result.Contents {
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
func Handler(ctx context.Context, event LambdaRequest) (LambdaResponse, error) {
	// Log that the function was invoked. These logs appear in CloudWatch Logs.
	log.Printf("Lambda function invoked. Checking S3 bucket for files...")

	// Get configuration from environment variables.
	// Environment variables are set in the Lambda function configuration.
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return LambdaResponse{
			StatusCode: 500,
			Message:    "S3_BUCKET_NAME environment variable is not set",
		}, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	// Get ECS configuration from environment variables
	ecsClusterName := os.Getenv("ECS_CLUSTER_NAME")
	ecsTaskDefinition := os.Getenv("ECS_TASK_DEFINITION")
	ecsSubnetIds := os.Getenv("ECS_SUBNET_IDS") // Comma-separated subnet IDs
	ecsSecurityGroupId := os.Getenv("ECS_SECURITY_GROUP_ID")

	log.Printf("Checking bucket: %s", bucketName)

	// Create an S3Checker instance to interact with S3.
	checker, err := NewS3Checker(bucketName)
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create S3 checker: %v", err),
		}, err
	}

	// Check for files in the S3 bucket.
	files, err := checker.CheckForFiles(ctx)
	if err != nil {
		return LambdaResponse{
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
		}

		// If ECS configuration is provided, trigger ECS tasks for each file
		if ecsClusterName != "" && ecsTaskDefinition != "" && ecsSubnetIds != "" && ecsSecurityGroupId != "" {
			log.Printf("Triggering ECS tasks to process %d file(s)...", len(files))

			// Load AWS config for ECS client
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				log.Printf("Warning: Failed to load AWS config for ECS: %v", err)
			} else {
				ecsClient := ecs.NewFromConfig(cfg)

				// Process each file by triggering an ECS task
				tasksTriggered := 0
				for _, file := range files {
					if err := triggerECSTask(ctx, ecsClient, ecsClusterName, ecsTaskDefinition, ecsSubnetIds, ecsSecurityGroupId, bucketName, file.Name); err != nil {
						log.Printf("Error triggering ECS task for file %s: %v", file.Name, err)
					} else {
						tasksTriggered++
						log.Printf("Successfully triggered ECS task for file: %s", file.Name)
					}
				}

				log.Printf("Triggered %d ECS task(s) out of %d file(s)", tasksTriggered, len(files))
			}
		} else {
			log.Println("ECS configuration not provided. Skipping ECS task triggering.")
			log.Println("To enable ECS processing, set: ECS_CLUSTER_NAME, ECS_TASK_DEFINITION, ECS_SUBNET_IDS, ECS_SECURITY_GROUP_ID")
		}
	} else {
		log.Println("No files found in the bucket.")
	}

	// Return a successful response with the file information.
	// Lambda will automatically convert this struct to JSON.
	return LambdaResponse{
		StatusCode: 200,
		Message:    fmt.Sprintf("Successfully checked bucket. Found %d file(s)", len(files)),
		FileCount:  len(files),
		Files:      files,
	}, nil
}

// triggerECSTask runs an ECS Fargate task to process a CSV file.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - ecsClient: ECS service client
//   - clusterName: Name of the ECS cluster
//   - taskDefinition: ARN or family name of the task definition
//   - subnetIds: Comma-separated subnet IDs where the task will run
//   - securityGroupId: Security group ID for the task
//   - bucketName: S3 bucket name
//   - fileName: Name of the file to process
//
// Returns: error if task couldn't be triggered
//
// How it works:
// 1. Parses subnet IDs from comma-separated string
// 2. Constructs input/output file paths in S3
// 3. Creates an ECS RunTask request with the file information
// 4. The ECS task will download the file, process it, and upload the result
func triggerECSTask(ctx context.Context, ecsClient *ecs.Client, clusterName, taskDefinition, subnetIds, securityGroupId, bucketName, fileName string) error {
	// Parse subnet IDs from comma-separated string
	// Example: "subnet-123,subnet-456" -> ["subnet-123", "subnet-456"]
	var subnetIdList []string
	if subnetIds != "" {
		// Split by comma and trim whitespace
		parts := strings.Split(subnetIds, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				subnetIdList = append(subnetIdList, trimmed)
			}
		}
	}

	if len(subnetIdList) == 0 {
		return fmt.Errorf("no valid subnet IDs provided")
	}

	// Create output filename (add "cleaned_" prefix)
	// Example: "data.csv" -> "cleaned_data.csv"
	outputFileName := fmt.Sprintf("cleaned_%s", fileName)

	// Build the command for the container
	// The container expects: -input <s3://bucket/file> -output <s3://bucket/output>
	// Note: The Go app currently expects local file paths, so we'll need to modify
	// the approach. For now, we'll pass the S3 paths as environment variables
	// and the container can download/upload using AWS CLI or SDK
	inputPath := fmt.Sprintf("s3://%s/%s", bucketName, fileName)
	outputPath := fmt.Sprintf("s3://%s/%s", bucketName, outputFileName)

	log.Printf("Triggering ECS task for file: %s -> %s", inputPath, outputPath)

	// Prepare the RunTask input
	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(clusterName),
		TaskDefinition: aws.String(taskDefinition),
		LaunchType:     types.LaunchTypeFargate, // Use Fargate (serverless)

		// Network configuration - required for Fargate
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        subnetIdList,
				SecurityGroups: []string{securityGroupId},
				AssignPublicIp: types.AssignPublicIpEnabled, // Needed for internet access (ECR, S3)
			},
		},

		// Override container command and environment variables
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name: aws.String("csv-handler"),
					// Set environment variables with file information
					Environment: []types.KeyValuePair{
						{
							Name:  aws.String("INPUT_FILE"),
							Value: aws.String(inputPath),
						},
						{
							Name:  aws.String("OUTPUT_FILE"),
							Value: aws.String(outputPath),
						},
						{
							Name:  aws.String("S3_BUCKET_NAME"),
							Value: aws.String(bucketName),
						},
					},
				},
			},
		},
	}

	// Call RunTask API to start the ECS task
	result, err := ecsClient.RunTask(ctx, runTaskInput)
	if err != nil {
		return fmt.Errorf("failed to run ECS task: %w", err)
	}

	// Log task information
	if len(result.Tasks) > 0 {
		log.Printf("ECS task started: %s (ARN: %s)", *result.Tasks[0].TaskArn, *result.Tasks[0].TaskArn)
	}

	return nil
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
