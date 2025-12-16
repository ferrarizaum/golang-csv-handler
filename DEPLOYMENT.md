# Deployment Guide

This guide walks you through deploying the complete solution step by step.

## Overview

This project has three main components:
1. **CLI Tool** (`main.go`): Command-line CSV handler (runs locally)
2. **Lambda Function** (`lambda/`): Serverless function that checks S3 for files and triggers ECS tasks
3. **ECS Fargate Container**: Processes CSV files (downloads, cleans, uploads back to S3)

## Architecture

```
EventBridge Scheduler (Cron)
    ↓ (triggers)
Lambda Function (checks S3 for CSV files)
    ↓ (triggers ECS task for each file)
ECS Fargate Container
    ↓ (downloads, processes, uploads)
S3 Bucket (stores cleaned CSV files)
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
go mod tidy  # Update dependencies (includes ECS SDK)
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
- Lambda function
- ECS cluster and task definition
- ECR repository (for Docker images)
- EventBridge schedule
- IAM roles and policies
- CloudWatch log groups

Review the output carefully!

### Step 6: Deploy All Infrastructure (Lambda + ECS)

```bash
terraform apply
```

Terraform will:
1. Show the plan again
2. Ask for confirmation (type `yes`)
3. Create all resources:
   - S3 bucket
   - Lambda function
   - ECS cluster and task definition
   - ECR repository (for Docker images)
   - EventBridge schedule
   - IAM roles and policies
   - CloudWatch log groups
4. Show outputs (bucket name, Lambda ARN, ECR URL, etc.)

**Expected time:** 2-3 minutes

**Note:** The ECS task definition is created, but tasks won't run until you push the Docker image (Step 7).

### Step 7: Build and Push Docker Image (for ECS)

After Terraform creates the ECR repository, build and push your Docker image:

**Step 1: Get your ECR repository URL and region**

Run these commands in your terminal (from the `terraform` directory):

```powershell
# Windows PowerShell - Get ECR URL and region from Terraform
$ECR_URL = terraform output -raw ecr_repository_url
$REGION = terraform output -raw aws_region

# Verify both are set
Write-Host "ECR URL: $ECR_URL"
Write-Host "Region: $REGION"
```

**If `$REGION` is still empty**, you can get it from your AWS CLI config:
```powershell
$REGION = aws configure get region
```

Or check your `terraform.tfvars` file and set it manually:
```powershell
$REGION = "us-east-1"  # Replace with the region from your terraform.tfvars file
```

**Step 2: Authenticate Docker with ECR**

First, verify your ECR URL format is correct:
```powershell
# Check the ECR URL format
Write-Host "ECR URL: $ECR_URL"
# Should look like: 123456789012.dkr.ecr.us-east-1.amazonaws.com/csv-handler-repo
# Must include the repository name at the end!
```

If the URL is missing the repository name (e.g., just `account.dkr.ecr.region.amazonaws.com`), you need to add it:
```powershell
# Get just the repository name from Terraform
$REPO_NAME = terraform output -raw ecs_service_name
$REPO_NAME = "$REPO_NAME-repo"  # ECR repo name format
$ECR_URL = "$ECR_URL/$REPO_NAME"
Write-Host "Full ECR URL: $ECR_URL"
```

Now authenticate:
```powershell
# Windows PowerShell - Authenticate with ECR
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_URL
```

**Important:** 
- Make sure you're in the `terraform` directory when running `terraform output`
- The ECR URL must include the repository name: `account.dkr.ecr.region.amazonaws.com/repo-name`
- If you get "400 Bad Request", the URL format is likely wrong (missing repo name)
- If `$ECR_URL` is empty, the docker login command will fail and try Docker Hub instead
-- // FIXED IT RUNNING POWERSHELL-safe version | if any docker command fail, echo the url and use it directly

**Build and push the image:**

```bash
# From project root directory
docker build -t csv-handler .

# Tag the image
# Linux/macOS/Git Bash:
docker tag csv-handler:latest $ECR_URL:latest

# Windows PowerShell:
docker tag csv-handler:latest $ECR_URL:latest


# Push to ECR
docker push $ECR_URL:latest
```

**Troubleshooting:**

**Error: "400 Bad Request" when logging in:**
- Verify the ECR URL includes the repository name: `account.dkr.ecr.region.amazonaws.com/repo-name`
- Check the repository exists: `aws ecr describe-repositories --region $REGION`
- Make sure you're using the correct region (must match where ECR was created)
- Try getting a fresh login token: `aws ecr get-login-password --region $REGION`

**Error: "unauthorized" error:**
- Make sure `$ECR_URL` is set correctly
- Verify the URL looks like: `123456789012.dkr.ecr.us-east-1.amazonaws.com/csv-handler-repo`
- Check your AWS credentials: `aws sts get-caller-identity`
- Verify you have ECR permissions

**Verify ECR repository exists:**
```powershell
# List all ECR repositories
aws ecr describe-repositories --region $REGION

# Check if your specific repo exists
aws ecr describe-repositories --repository-names csv-handler-repo --region $REGION
```

**Important Notes:**
- This uses your **AWS CLI credentials** (the ones you set up with `aws configure` in Step 1)
- The username is always **"AWS"** (not your email address!)
- The password is a temporary token that AWS CLI generates automatically
- Replace `<your-region>` with your AWS region (e.g., `us-east-1`)
- This step is required before ECS tasks can run. Lambda will work fine without it.

### Step 8: Verify Deployment

#### Check Lambda Function
```bash
aws lambda get-function --function-name s3-file-checker
```

#### Check S3 Bucket
```bash
aws s3 ls s3://<your-bucket-name>
```

#### Check ECS Cluster
```bash
aws ecs describe-clusters --clusters csv-handler-cluster
```

#### Check ECR Repository
```bash
aws ecr describe-repositories --repository-names csv-handler-repo
```

#### Check EventBridge Schedule
```bash
aws scheduler get-schedule --name s3-file-checker-schedule --group-name default
```

### Step 9: Test the Lambda

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

### Step 10: Test ECS Processing

1. Upload a CSV file to S3:
   ```bash
   echo "name,email,age
   John,john@example.com,30
   Jane,jane@example.com,25" > test.csv
   aws s3 cp test.csv s3://<your-bucket-name>/test.csv
   ```

2. Manually trigger Lambda (or wait for schedule):
   ```bash
   aws lambda invoke --function-name s3-file-checker --payload '{}' response.json
   ```

3. Lambda will automatically trigger an ECS task to process the file

4. Check ECS logs:
   ```bash
   aws logs tail /ecs/csv-handler --follow
   ```

5. Verify cleaned file appears in S3:
   ```bash
   aws s3 ls s3://<your-bucket-name>/
   # You should see "cleaned_test.csv"
   ```

### Step 11: Wait for Scheduled Execution

The Lambda will automatically run based on your schedule (e.g., every hour) and trigger ECS tasks for any CSV files found.

Check CloudWatch Logs:
- Lambda logs: `/aws/lambda/s3-file-checker`
- ECS logs: `/ecs/csv-handler`

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
- ECS cluster and task definitions
- ECR repository (and all Docker images)
- EventBridge schedule
- IAM roles
- CloudWatch logs

## How ECS Works

The Lambda function automatically triggers ECS Fargate tasks to process CSV files:

1. **Lambda** finds CSV files in S3
2. For each file, **Lambda** triggers an **ECS Fargate task**
3. **ECS task** downloads file from S3, processes it with your Go CSV handler, uploads cleaned version back
4. All logs go to CloudWatch

## ECS Configuration

Edit `terraform/terraform.tfvars` to customize ECS:
- `ecs_service_name` - Name for ECS resources (default: "csv-handler")
- `ecs_task_cpu` - CPU units: 256 (0.25 vCPU), 512 (0.5 vCPU), 1024 (1 vCPU)
- `ecs_task_memory` - Memory in MB (must match CPU limits)

### ECS Troubleshooting

**Task fails to start:**
- Check CloudWatch logs: `/ecs/csv-handler`
- Verify Docker image was pushed successfully
- Check IAM roles have correct permissions

**Container can't access S3:**
- Verify task role has S3 permissions
- Check security group allows outbound traffic
- Ensure subnet has internet access

**Lambda can't trigger tasks:**
- Check Lambda IAM role has `ecs:RunTask` permission
- Verify ECS cluster name and task definition are correct
- Check Lambda environment variables are set

## Next Steps

1. **Monitor costs**: Use AWS Cost Explorer to track ECS usage
2. **Add retry logic**: Handle failed ECS tasks
3. **Set up alarms**: CloudWatch alarms for task failures
4. **Optimize resources**: Adjust CPU/memory based on file sizes

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

