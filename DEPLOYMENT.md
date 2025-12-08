# Deployment Guide - AWS Lambda S3 File Checker

This guide walks you through deploying the complete solution step by step.

## Overview

This project has two main components:
1. **CLI Tool** (`main.go`): Command-line CSV handler (runs locally)
2. **Lambda Function** (`lambda/`): Serverless function that checks S3 for files (runs in AWS)

## Architecture

```
EventBridge Scheduler (Cron)
    ↓ (triggers)
Lambda Function
    ↓ (checks)
S3 Bucket (CSV files)
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
make build
```

This creates `lambda/bootstrap` (or `bootstrap` in the lambda directory) - the executable that Lambda will run.

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
schedule_expression = "rate(1 hour)"  # Check every hour
```

**Important:** S3 bucket names must be globally unique. Use random numbers/letters.

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
- S3 bucket
- IAM roles and policies
- Lambda function
- EventBridge schedule
- CloudWatch log group

Review the output carefully!

### Step 6: Deploy Infrastructure

```bash
terraform apply
```

Terraform will:
1. Show the plan again
2. Ask for confirmation (type `yes`)
3. Create all resources
4. Show outputs (bucket name, Lambda ARN, etc.)

**Expected time:** 1-2 minutes

### Step 7: Verify Deployment

#### Check Lambda Function
```bash
aws lambda get-function --function-name s3-file-checker
```

#### Check S3 Bucket
```bash
aws s3 ls s3://<your-bucket-name>
```

#### Check EventBridge Schedule
```bash
aws scheduler get-schedule --name s3-file-checker-schedule --group-name default
```

### Step 8: Test the Lambda

#### Upload a Test File
```bash
echo "name,email,age
John,john@example.com,30" > test.csv

aws s3 cp test.csv s3://<your-bucket-name>/test.csv
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
      "name": "test.csv",
      "size": 45,
      "lastModified": "2024-01-15 10:30:00"
    }
  ]
}
```

#### View Logs
```bash
aws logs tail /aws/lambda/s3-file-checker --follow
```

### Step 9: Wait for Scheduled Execution

The Lambda will automatically run based on your schedule (e.g., every hour).

Check CloudWatch Logs to see scheduled executions:
1. AWS Console → CloudWatch → Log Groups
2. Find `/aws/lambda/s3-file-checker`
3. View recent log streams

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
```

Then:
```bash
cd terraform
terraform apply
```

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

### Lambda Error: "Access Denied"
- Check IAM role has S3 permissions
- Verify bucket name is correct

### No logs appearing
- Check CloudWatch Log Group exists
- Verify Lambda has logging permissions
- Wait a few minutes (logs can be delayed)

### Schedule not triggering
- Check schedule is ENABLED
- Verify IAM role has invoke permission
- Check schedule expression syntax

## Cleanup

To remove all resources:

```bash
cd terraform
terraform destroy
```

**Warning:** This deletes:
- S3 bucket and all files
- Lambda function
- EventBridge schedule
- IAM roles
- CloudWatch logs

## Next Steps

1. **Process CSV files**: Extend Lambda to clean CSV files found in S3
2. **Move files**: Move processed files to a different S3 folder
3. **Send notifications**: Use SNS to email when files are found
4. **Add error handling**: Handle edge cases and errors gracefully
5. **Add tests**: Write unit tests for Lambda function

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

