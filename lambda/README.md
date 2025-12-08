# Lambda Function - S3 File Checker

This Lambda function checks for files in an S3 bucket when triggered by EventBridge Scheduler.

## What This Lambda Does

1. **Reads Configuration**: Gets the S3 bucket name from environment variables
2. **Lists Files**: Uses AWS SDK to list all objects in the S3 bucket
3. **Returns Results**: Returns information about found files (name, size, last modified)

## Building the Lambda Function

Lambda functions need to be compiled for Linux (Amazon Linux 2).

### Windows

**Option 1: Using Make (recommended)**
```powershell
# From the lambda directory
make build
```

**Option 2: Using PowerShell script**
```powershell
# From the lambda directory
.\build.ps1
```

**Option 3: Using batch file**
```cmd
# From the lambda directory
build.bat
```

### Linux/macOS/Git Bash

```bash
# From the lambda directory
make build
```

### Manual Build (all platforms)

**Windows (PowerShell):**
```powershell
# From the lambda directory
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -o bootstrap .
```

**Windows (CMD):**
```cmd
# From the lambda directory
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o bootstrap .
```

**Unix-like (Linux/macOS/Git Bash):**
```bash
# From the lambda directory
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap .
```

## Local Testing

You can test the Lambda function locally using the AWS SAM CLI or by creating a test event:

```bash
# Install AWS SAM CLI first: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html

# Test locally
sam local invoke S3CheckerFunction --event event.json
```

## Environment Variables

- `S3_BUCKET_NAME`: The name of the S3 bucket to check (set by Terraform)

## AWS Permissions Required

The Lambda function needs:
- `s3:ListBucket` - To list objects in the bucket
- `s3:GetObject` - To read object metadata
- `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents` - To write logs

These are configured in the Terraform IAM role.

## Deployment

The Lambda function is deployed automatically by Terraform. See the `terraform/` directory for deployment instructions.

