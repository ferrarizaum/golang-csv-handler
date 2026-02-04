# CSV Handler

A Go application for cleaning and processing CSV files, with support for both local file processing and AWS Lambda deployment for S3-based processing.

## Features

- Remove duplicate rows
- Trim whitespace from fields
- Filter non-printable characters
- Normalize line endings
- Escape unescaped quotes
- Skip empty rows

## Project Structure

```
.
├── cmd/
│   ├── local/          # CLI application for local file processing
│   └── lambda/         # AWS Lambda handler for S3 processing
├── internal/
│   ├── csv/            # CSV cleaning and processing logic
│   └── s3/             # S3 client and operations
├── lambda/             # Lambda build scripts and deployment
└── terraform/          # Infrastructure as Code
```

## Usage

### Local Processing

Process CSV files on your local machine:

```bash
go run cmd/local/main.go -input dirty.csv -output clean.csv
```

Or build and run:

```bash
go build -o csv-cleaner cmd/local/main.go
./csv-cleaner -input dirty.csv -output clean.csv
```

### AWS Lambda Deployment

1. Build the Lambda deployment package:

```powershell
cd lambda
./build.ps1
```

2. Deploy using Terraform (see `terraform/README.md`)

The Lambda function will:
- Monitor the `input/` folder in your S3 bucket
- Process new CSV files automatically
- Save cleaned files to `output/` folder
- Archive originals to `archive/` folder
- Delete processed files from `input/`

## Environment Variables (Lambda)

- `S3_BUCKET_NAME`: Name of the S3 bucket to monitor

## Development

### Requirements

- Go 1.25.2 or later
- AWS credentials (for Lambda deployment)
- Terraform (for infrastructure deployment)

### Building

Local CLI:
```bash
go build -o csv-cleaner cmd/local/main.go
```

Lambda function:
```bash
cd lambda
./build.ps1
```

### Testing

Run tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## Best Practices Applied

- **Clear package structure**: Separation of concerns with `internal/` packages
- **Dependency injection**: Easier testing and maintenance
- **Error wrapping**: Context-aware error messages
- **Constants**: No magic strings or numbers
- **Godoc comments**: Proper documentation for exported functions
- **Input validation**: Early validation of inputs
- **Resource cleanup**: Proper use of `defer` for resource management
- **Structured logging**: Clear, actionable log messages
- **Idiomatic Go**: Following standard Go conventions and patterns

## License

MIT
