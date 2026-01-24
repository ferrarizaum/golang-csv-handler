# Deployment Guide

This guide walks you through deploying the complete solution step by step.

## Overview

This project has two main components:
1. **CLI Tool** (`main.go`): Command-line CSV handler (runs locally)
2. **Lambda Function** (`lambda/`): Serverless function that checks S3 for files and processes them directly (downloads, cleans, uploads back to S3)

## Architecture

```
EventBridge Scheduler (Cron)
    ↓ (triggers on schedule)
Lambda Function
    ↓ (checks S3 input/ folder for CSV files)
    ↓ (downloads, cleans, and processes each file)
    ↓ (uploads cleaned file to output/ folder)
    ↓ (deletes original from input/ folder)
S3 Bucket
    ├── input/  (raw CSV files)
    └── output/ (cleaned CSV files with _cleaned.csv suffix)
    ↓ (logs)
CloudWatch Logs
```

## Step-by-Step Deployment

### Step 1: Prerequisites Setup

#### Install AWS CLI
```bash
# Windows (Chocolatey)
choco install awscli

# macOS
brew install awscli

# Linux
sudo apt-get install awscli
```

#### Configure AWS Credentials
```bash
aws configure
```

Enter:
- **AWS Access Key ID**: Your AWS access key
- **AWS Secret Access Key**: Your AWS secret key
- **Default region**: e.g., `us-east-1`
- **Default output format**: `json`

> **Getting AWS Credentials:**
> 1. Go to AWS Console → IAM → Users
> 2. Create a user or select existing
> 3. Attach policy: `AdministratorAccess` (for learning) or create custom policy
> 4. Go to "Security credentials" tab
> 5. Create access key

#### Install Terraform
```bash
# Windows (Chocolatey)
choco install terraform

# macOS
brew install terraform

# Linux
sudo apt-get install terraform
```

Verify installation:
```bash
terraform version
```

### Step 2: Build Lambda Function

The Lambda function must be compiled for Linux (Amazon Linux 2):

```bash
# From project root
cd lambda
go mod tidy  # Update dependencies
make build
```

This creates `lambda/bootstrap` - the executable that Lambda will run.

**What's happening:**
- `GOOS=linux`: Compile for Linux OS
- `GOARCH=amd64`: 64-bit architecture
- `CGO_ENABLED=0`: Static binary (no external dependencies)

### Step 3: Configure Terraform Variables

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:
```hcl
aws_region = "us-east-1"
environment = "dev"
s3_bucket_name = "my-csv-files-bucket-12345"  # MUST be globally unique!
lambda_function_name = "s3-file-checker"
schedule_expression = "rate(5 minutes)"  # Check every 5 minutes
```

**Important:** 
- S3 bucket names must be globally unique across all AWS accounts. Use random numbers/letters.
- S3 bucket names can only contain lowercase letters, numbers, dots, and hyphens (no underscores!)
- The schedule expression can be:
  - `rate(5 minutes)` - Every 5 minutes
  - `rate(1 hour)` - Every hour
  - `rate(1 day)` - Every day
  - `cron(0 9 * * ? *)` - Every day at 9:00 AM UTC

### Step 4: Initialize Terraform

```bash
cd terraform
terraform init
```

This downloads:
- AWS provider (to create AWS resources)
- Archive provider (to create ZIP files)

### Step 5: Review Deployment Plan

```bash
terraform plan
```

This shows what will be created:
- S3 bucket with input/ and output/ folders
- Lambda function
- EventBridge schedule
- IAM roles and policies
- CloudWatch log groups

Review the output carefully!

### Step 6: Deploy All Infrastructure

```bash
terraform apply
```

Terraform will:
1. Show the plan again
2. Ask for confirmation (type `yes`)
3. Create all resources:
   - S3 bucket with input/ and output/ folders
   - Lambda function
   - EventBridge schedule
   - IAM roles and policies
   - CloudWatch log groups
4. Show outputs (bucket name, Lambda ARN, etc.)

**Expected time:** 1-2 minutes

### Step 7: Verify Deployment

#### Check Lambda Function
```bash
aws lambda get-function --function-name s3-file-checker
```

#### Check S3 Bucket and Folders
```bash
# List bucket contents
aws s3 ls s3://<your-bucket-name>

# Should show input/ and output/ folders
```

#### Check EventBridge Schedule
```bash
aws scheduler get-schedule --name s3-file-checker-schedule --group-name default
```

### Step 8: Test the Lambda

#### Upload a Test File
```bash
# Create a test CSV file
echo "name,email,age
John,john@example.com,30
Jane,jane@example.com,25" > test.csv

# Upload to the input folder
aws s3 cp test.csv s3://<your-bucket-name>/input/test.csv
```

#### Manually Invoke Lambda
```bash
aws lambda invoke \
  --function-name s3-file-checker \
  --payload '{}' \
  response.json

cat response.json
```

You should see:
```json
{
  "statusCode": 200,
  "message": "Successfully checked bucket. Found 1 file(s)",
  "fileCount": 1,
  "files": [
    {
      "name": "input/test.csv",
      "size": 67,
      "lastModified": "2026-01-24 10:30:00"
    }
  ]
}
```

#### View Lambda Logs
```bash
aws logs tail /aws/lambda/s3-file-checker --follow
```

You should see logs showing:
- File found in input/ folder
- Processing file
- CSV content being read
- File uploaded to output/ folder
- Original file deleted from input/ folder

#### Verify Cleaned File
```bash
# Check output folder for cleaned file
aws s3 ls s3://<your-bucket-name>/output/

# Download and view the cleaned file
aws s3 cp s3://<your-bucket-name>/output/test.csv_cleaned.csv cleaned_test.csv
cat cleaned_test.csv
```

#### Verify Original File Deleted
```bash
# Check that input folder is now empty
aws s3 ls s3://<your-bucket-name>/input/
```

### Step 9: Wait for Scheduled Execution

The Lambda will automatically run based on your schedule (e.g., every 5 minutes) and process any CSV files found in the input/ folder.

Check CloudWatch Logs:
- Lambda logs: `/aws/lambda/s3-file-checker`

## Understanding AWS Services

### AWS Lambda
- **What it is**: Serverless compute - you upload code, AWS runs it
- **Cost**: Pay per request and compute time
- **Use case**: Run code without managing servers

### EventBridge Scheduler
- **What it is**: Serverless cron service
- **Cost**: Free for first 14M invocations/month
- **Use case**: Trigger Lambda on a schedule

### S3 (Simple Storage Service)
- **What it is**: Object storage (like a file system in the cloud)
- **Cost**: ~$0.023/GB storage + requests
- **Use case**: Store files (CSV files in our case)

### CloudWatch
- **What it is**: Monitoring and logging service
- **Cost**: ~$0.50/GB logs ingested
- **Use case**: View Lambda logs and metrics

### IAM (Identity and Access Management)
- **What it is**: Controls who can do what in AWS
- **Cost**: Free
- **Use case**: Give Lambda permission to read S3

## Updating the Lambda

If you change the Lambda code:

1. **Rebuild:**
   ```bash
   cd lambda
   make build
   ```

2. **Redeploy:**
   ```bash
   cd terraform
   terraform apply
   ```

Terraform detects the file change and updates the Lambda.

## Changing the Schedule

Edit `terraform/terraform.tfvars`:
```hcl
schedule_expression = "rate(5 minutes)"  # Check every 5 minutes
# OR
schedule_expression = "rate(1 hour)"     # Check every hour
# OR
schedule_expression = "cron(0 9 * * ? *)" # Every day at 9:00 AM UTC
```

Then apply the changes:
```bash
cd terraform
terraform apply
```

The schedule will be updated without affecting your S3 bucket or files.

## Monitoring

### View Lambda Metrics
- AWS Console → Lambda → Functions → s3-file-checker → Monitoring

### View Logs
```bash
aws logs tail /aws/lambda/s3-file-checker --follow
```

### Set Up Alarms (Optional)
Create CloudWatch alarms to notify you of errors or when files are found.

## Troubleshooting

### Lambda Error: "S3_BUCKET_NAME environment variable is not set"
- Check Terraform applied successfully
- Verify environment variable in Lambda console
- Run `terraform apply` again to ensure all resources are created

### Lambda Error: "Access Denied" when accessing S3
- Check IAM role has S3 permissions (ListBucket, GetObject, PutObject, DeleteObject)
- Verify bucket name is correct
- Check the bucket exists: `aws s3 ls s3://<your-bucket-name>`

### Lambda Error: "failed to parse cleaned CSV"
- Check CloudWatch logs for specific CSV parsing errors
- The cleaner handles most issues, but some files may need manual review
- Try downloading the file and checking its format

### No logs appearing in CloudWatch
- Check CloudWatch Log Group exists: `/aws/lambda/s3-file-checker`
- Verify Lambda has logging permissions
- Wait a few minutes (logs can be delayed)
- Check IAM role has `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents` permissions

### Schedule not triggering Lambda
- Check schedule is ENABLED: `aws scheduler get-schedule --name s3-file-checker-schedule --group-name default`
- Verify IAM role has invoke permission
- Check schedule expression syntax in `terraform.tfvars`
- View schedule history in AWS Console → EventBridge → Schedules

### Files not being deleted from input/ folder
- Check Lambda logs for deletion errors
- Verify IAM role has `s3:DeleteObject` permission
- Check if Lambda completed successfully (no timeout)

## Cleanup

To remove all resources:

```bash
cd terraform
terraform destroy
```

**Warning:** This deletes:
- S3 bucket and all files (including input/ and output/ folders)
- Lambda function
- EventBridge schedule
- IAM roles and policies
- CloudWatch logs

## How CSV Processing Works

The Lambda function directly processes CSV files:

1. **EventBridge Scheduler** triggers Lambda on schedule (e.g., every 5 minutes)
2. **Lambda** checks the S3 bucket's `input/` folder for CSV files
3. For each CSV file found:
   - Downloads the file from S3
   - Cleans the data:
     - Removes or escapes problematic characters
     - Normalizes line endings (converts `\r\n` and `\r` to `\n`)
     - Handles malformed quotes
     - Removes empty lines
     - Removes non-printable characters
   - Parses the CSV with lenient settings (LazyQuotes, TrimLeadingSpace)
   - Uploads cleaned file to `output/` folder with `_cleaned.csv` suffix
   - Deletes the original file from `input/` folder
4. All operations are logged to CloudWatch

## Lambda Configuration

The Lambda function has the following settings (configured in `terraform/main.tf`):
- **Runtime**: `provided.al2023` (Go custom runtime)
- **Architecture**: `x86_64` (matches the amd64 build)
- **Timeout**: 30 seconds
- **Memory**: Default (128 MB)
- **Environment Variables**:
  - `S3_BUCKET_NAME` - The S3 bucket to check

### Lambda Troubleshooting

**Lambda timeout:**
- Increase timeout in `terraform/main.tf` (currently 30 seconds)
- Check CloudWatch logs for processing time
- Consider splitting large files

**CSV parsing errors:**
- Check CloudWatch logs for specific error messages
- Verify CSV file format is valid
- The cleaner handles most common issues automatically

**Files not being processed:**
- Verify files are in the `input/` folder (not root)
- Check Lambda IAM role has S3 permissions
- Check EventBridge schedule is enabled
- Verify Lambda environment variables are set

## Next Steps

1. **Monitor costs**: Use AWS Cost Explorer to track Lambda usage
2. **Add retry logic**: Handle failed processing attempts
3. **Set up alarms**: CloudWatch alarms for Lambda errors
4. **Optimize timeout**: Adjust based on actual file sizes

## Cost Optimization

- **Reduce schedule frequency**: Check less often if not needed
- **Set log retention**: Keep logs for shorter periods (7 days default)
- **Use S3 lifecycle policies**: Move old files to cheaper storage
- **Monitor usage**: Use AWS Cost Explorer to track spending

## Security Best Practices

1. **Least privilege**: IAM roles only have needed permissions
2. **No public access**: S3 bucket blocks public access
3. **Encryption**: Enable S3 encryption (can be added to Terraform)
4. **Secrets**: Use AWS Secrets Manager for sensitive data (not needed here)

## Support

- **AWS Documentation**: https://docs.aws.amazon.com/
- **Terraform AWS Provider**: https://registry.terraform.io/providers/hashicorp/aws/latest/docs
- **Lambda Go Runtime**: https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html

