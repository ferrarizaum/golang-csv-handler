package checkforfiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Checker) CheckForFiles(ctx context.Context) ([]FileInfo, error) {
	// ListObjectsV2Input is the input structure for listing objects in S3.
	// We specify the bucket name we want to list objects from.
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket), // The bucket to list objects from
		Prefix: aws.String("input/"), // The prefix to filter by
		// MaxKeys: You can limit the number of results (optional)
		// Prefix: You can filter by prefix (e.g., "input/") (optional)
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
