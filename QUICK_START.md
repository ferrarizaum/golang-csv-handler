# Quick Start Guide

## Installation

Clone the repository and build:

```bash
git clone <repository-url>
cd golang-csv-handler
go mod download
make build
```

## Local Usage

### Using Go Run
```bash
go run cmd/local/main.go -input dirty.csv -output clean.csv
```

### Using Compiled Binary
```bash
# Build first
make build-local

# Run
./bin/csv-cleaner -input dirty.csv -output clean.csv
```

### Help
```bash
go run cmd/local/main.go -h
```

## AWS Lambda Deployment

### 1. Build Lambda Package
```bash
cd lambda
./build.ps1
```

This creates `lambda/function.zip` ready for deployment.

### 2. Deploy with Terraform

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
terraform init
terraform plan
terraform apply
```

### 3. Configure Environment

Set the `S3_BUCKET_NAME` environment variable in your Lambda function configuration.

### 4. Test

Upload a CSV file to the `input/` folder in your S3 bucket. The Lambda will:
- Process it automatically (via EventBridge schedule)
- Save cleaned version to `output/`
- Archive original to `archive/`
- Delete from `input/`

## What Gets Cleaned?

- ✅ Duplicate rows removed
- ✅ Empty rows removed
- ✅ Whitespace trimmed
- ✅ Non-printable characters filtered
- ✅ Line endings normalized
- ✅ Unescaped quotes fixed

## Project Structure

```
golang-csv-handler/
├── cmd/
│   ├── local/          # CLI application
│   └── lambda/         # Lambda handler
├── internal/
│   ├── csv/           # CSV cleaning logic
│   └── s3/            # S3 operations
├── lambda/            # Lambda build scripts
├── terraform/         # Infrastructure code
└── bin/              # Compiled binaries
```

## Common Commands

```bash
# Build everything
make build

# Run tests
make test

# Clean build artifacts
make clean

# Format code
make fmt

# Run example
make example
```

## Troubleshooting

### "Command not found"
Make sure you've built the binary:
```bash
make build-local
```

### "File not found"
Check the file path is correct:
```bash
ls test-data-dirty.csv
```

### Lambda not processing files
1. Check CloudWatch logs
2. Verify S3_BUCKET_NAME environment variable
3. Check IAM permissions
4. Verify EventBridge rule is enabled

## Next Steps

- Read `README.md` for detailed documentation
- Read `REFACTORING.md` to understand the code structure
- Check `terraform/README.md` for infrastructure details
- Add tests in `internal/csv` and `internal/s3` packages

## Support

For issues, questions, or contributions, please refer to the main README.md.
