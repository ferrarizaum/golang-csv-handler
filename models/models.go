package models

import "github.com/aws/aws-sdk-go-v2/service/s3"

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

type S3Checker struct {
	s3Client *s3.Client // S3 client used to make API calls to S3
	bucket   string     // Name of the S3 bucket to check
}
