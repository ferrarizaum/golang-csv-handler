# Golang CSV Handler

A Go-based serverless project for automated CSV file processing using AWS Lambda and S3.

## Project Overview

This project consists of two main components:

1. **CLI Tool** (`main.go`) - Command-line CSV handler for local file processing
2. **Lambda Function** (`lambda/`) - Serverless function that checks S3 buckets for CSV files on a schedule and processes them directly (cleans data, removes duplicates)

## Quick Start

### 1. Build Lambda
```bash
cd lambda
go mod tidy
make build
```

### 2. Configure Terraform
```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars - change s3_bucket_name (must be globally unique!)
```

### 3. Deploy Infrastructure
```bash
terraform init
terraform plan
terraform apply
```

### 4. Test
```bash
# Upload a CSV file to the input folder
aws s3 cp test.csv s3://<your-bucket>/input/test.csv

# Lambda will process it automatically on schedule, or manually invoke:
aws lambda invoke --function-name s3-file-checker response.json

# Check the output folder for the cleaned file
aws s3 ls s3://<your-bucket>/output/
```

## Features

- **Automated Processing**: EventBridge Scheduler triggers Lambda on a configurable schedule
- **S3 Integration**: Monitors `input/` folder, processes files, saves to `output/` folder
- **Robust CSV Cleaning**: Handles malformed CSVs, bad quotes, line endings, and non-printable characters
- **Serverless Architecture**: No servers to manage, scales automatically
- **Cost Effective**: Pay only for what you use (Lambda execution time)
- **CloudWatch Logging**: Complete visibility into processing with detailed logs
- **Infrastructure as Code**: Terraform manages all AWS resources

## Prerequisites

- **Go** 1.21+ ([Install Go](https://golang.org/doc/install))
- **AWS CLI** ([Install AWS CLI](https://aws.amazon.com/cli/))
- **Terraform** ([Install Terraform](https://www.terraform.io/downloads))
- **AWS Account** with appropriate permissions

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd golang-csv-handler
```

### 2. Configure AWS Credentials

```bash
aws configure
```

Enter your AWS Access Key ID, Secret Access Key, region, and output format.

### 3. Configure Terraform Variables

**Important:** The `terraform.tfvars` file contains your actual deployment configuration and is **not committed to git** for security reasons.

Create your configuration file from the example:

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and update the values, especially:
- `s3_bucket_name` - Must be globally unique (use random numbers/letters)
- `aws_region` - Your preferred AWS region
- `schedule_expression` - How often Lambda should run

**Note:** S3 bucket names can only contain lowercase letters, numbers, dots, and hyphens (no underscores!).

### 4. Build the Lambda Function

```bash
# Windows
cd lambda
make build

# Linux/macOS
cd lambda
make build
```

This creates the `bootstrap` binary required for Lambda deployment.

### 5. Deploy to AWS

```bash
cd terraform
terraform init    # Download providers
terraform plan    # Review what will be created
terraform apply   # Deploy (type 'yes' to confirm)
```

## Architecture

```
EventBridge Scheduler
    ↓ (triggers on schedule)
Lambda Function
    ↓ (checks S3 input/ folder for CSV files)
    ↓ (downloads, cleans, and processes each file)
    ↓ (uploads cleaned file to output/ folder)
    ↓ (deletes original from input/ folder)
S3 Bucket
    ├── input/  (raw CSV files)
    └── output/ (cleaned CSV files)
```

## Project Structure

```
golang-csv-handler/
├── main.go              # CLI tool for local CSV processing
├── lambda/              # Lambda function code
│   ├── main.go         # Lambda handler (checks S3 and processes CSV files)
│   └── Makefile        # Build automation
└── terraform/          # Infrastructure as Code
    ├── main.tf         # All AWS resources (Lambda, S3, EventBridge)
    ├── variables.tf    # Configuration variables
    ├── outputs.tf      # Output values
    └── terraform.tfvars.example  # Configuration template
```

## Documentation

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Detailed deployment guide with troubleshooting

## Cleanup

To remove all AWS resources:

```bash
cd terraform
terraform destroy
```

This will delete:
- S3 bucket and all files (including input/ and output/ folders)
- Lambda function
- EventBridge schedule
- IAM roles and policies
- CloudWatch logs

## Monitoring

### View Lambda Logs

```bash
# Tail logs in real-time
aws logs tail /aws/lambda/s3-file-checker --follow

# View recent logs
aws logs tail /aws/lambda/s3-file-checker --since 1h
```

### Check Lambda Metrics

```bash
# Get function details
aws lambda get-function --function-name s3-file-checker

# View recent invocations
aws lambda get-function --function-name s3-file-checker --query 'Configuration.[LastModified,LastUpdateStatus]'
```

### Monitor S3 Bucket

```bash
# List files in input folder
aws s3 ls s3://<your-bucket>/input/

# List files in output folder
aws s3 ls s3://<your-bucket>/output/

# Check bucket size
aws s3 ls s3://<your-bucket> --recursive --summarize
```

## Cost Estimation

This solution is very cost-effective:

- **Lambda**: Free tier includes 1M requests/month and 400,000 GB-seconds of compute
  - After free tier: ~$0.20 per 1M requests
  - Typical cost: < $1/month for small workloads
- **S3**: $0.023/GB storage + $0.0004 per 1,000 GET requests
  - Typical cost: < $0.50/month for small files
- **EventBridge Scheduler**: Free for first 14M invocations/month
- **CloudWatch Logs**: $0.50/GB ingested
  - Typical cost: < $0.10/month with 7-day retention

**Example**: Processing 1,000 CSV files per month (100KB each):
- Lambda: ~$0.01
- S3: ~$0.05
- Total: **< $0.10/month**

## Development

### Local CLI Tool

Run the CSV handler locally:

```bash
go run main.go -input input.csv -output output.csv
```

### Lambda Function

Build the Lambda function:

```bash
cd lambda
make build
```

Test locally (requires AWS SAM CLI):

```bash
sam local invoke S3CheckerFunction --event event.json
```

## Security Notes

The following files are **NOT committed to git** (see `.gitignore`):

- `terraform/terraform.tfstate` - Contains sensitive AWS resource information
- `terraform/terraform.tfstate.backup` - Backup of state file
- `terraform/terraform.tfvars` - Contains your actual deployment values
- `lambda/bootstrap` - Compiled binary (regenerate with `make build`)
- `terraform/lambda_function.zip` - Generated deployment package

Always use `terraform.tfvars.example` as a template and never commit your actual `terraform.tfvars` file.

## How It Works

1. **EventBridge Scheduler** triggers the Lambda function on a schedule (e.g., every 5 minutes)
2. **Lambda function** checks the S3 bucket's `input/` folder for CSV files
3. For each CSV file found:
   - Downloads the file from S3
   - Cleans the data (removes problematic characters, normalizes line endings, handles malformed quotes)
   - Parses the CSV with lenient settings
   - Uploads the cleaned file to the `output/` folder with `_cleaned.csv` suffix
   - Deletes the original file from the `input/` folder
4. All operations are logged to CloudWatch for monitoring

## CSV Cleaning Features

The Lambda function includes robust CSV cleaning capabilities:

- **Line Ending Normalization**: Converts `\r\n` and `\r` to `\n`
- **Empty Line Removal**: Strips out blank lines
- **Quote Handling**: Fixes malformed quotes in CSV fields
- **Character Cleaning**: Removes non-printable characters (except tabs)
- **Lenient Parsing**: Uses `LazyQuotes` and `TrimLeadingSpace` for flexible parsing
- **Variable Fields**: Handles CSV files with inconsistent column counts

### File Processing Flow

1. Upload `myfile.csv` to `s3://your-bucket/input/myfile.csv`
2. Lambda processes it on the next scheduled run
3. Cleaned file appears at `s3://your-bucket/output/myfile.csv_cleaned.csv`
4. Original file is deleted from `input/` folder
5. All operations logged to CloudWatch: `/aws/lambda/s3-file-checker`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license here]

## Support

For issues and questions:
- Check the [DEPLOYMENT.md](./DEPLOYMENT.md) troubleshooting section
- Review AWS documentation for Lambda and S3
- Check Terraform AWS provider documentation
