#!/bin/bash
# Entrypoint script for the CSV handler container
# This script downloads files from S3, processes them, and uploads results back

set -e  # Exit on any error

# Get S3 paths from environment variables
INPUT_FILE="${INPUT_FILE:-}"
OUTPUT_FILE="${OUTPUT_FILE:-}"
S3_BUCKET_NAME="${S3_BUCKET_NAME:-}"

# Local file paths
LOCAL_INPUT="/tmp/input.csv"
LOCAL_OUTPUT="/tmp/output.csv"

# Check if INPUT_FILE is provided
if [ -z "$INPUT_FILE" ]; then
    echo "Error: INPUT_FILE environment variable is not set"
    echo "Expected format: s3://bucket-name/file.csv"
    exit 1
fi

# Check if OUTPUT_FILE is provided
if [ -z "$OUTPUT_FILE" ]; then
    echo "Error: OUTPUT_FILE environment variable is not set"
    echo "Expected format: s3://bucket-name/output.csv"
    exit 1
fi

echo "=========================================="
echo "CSV Handler - Processing file"
echo "=========================================="
echo "Input:  $INPUT_FILE"
echo "Output: $OUTPUT_FILE"
echo ""

# Download the input file from S3
echo "Step 1: Downloading file from S3..."
if ! aws s3 cp "$INPUT_FILE" "$LOCAL_INPUT"; then
    echo "Error: Failed to download file from S3"
    exit 1
fi
echo "✓ File downloaded successfully"
echo ""

# Process the CSV file using the Go application
echo "Step 2: Processing CSV file..."
if ! /app/csv-handler -input "$LOCAL_INPUT" -output "$LOCAL_OUTPUT"; then
    echo "Error: Failed to process CSV file"
    exit 1
fi
echo "✓ CSV file processed successfully"
echo ""

# Upload the result back to S3
echo "Step 3: Uploading result to S3..."
if ! aws s3 cp "$LOCAL_OUTPUT" "$OUTPUT_FILE"; then
    echo "Error: Failed to upload file to S3"
    exit 1
fi
echo "✓ Result uploaded successfully"
echo ""

echo "=========================================="
echo "Processing completed successfully!"
echo "=========================================="
