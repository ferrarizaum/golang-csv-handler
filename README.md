# Golang CSV Handler

A Go-based project for handling CSV files with a local CLI tool, AWS Lambda function, and ECS Fargate container for automated processing.

## Project Overview

This project consists of three main components:

1. **CLI Tool** (`main.go`) - Command-line CSV handler for local file processing
2. **Lambda Function** (`lambda/`) - Serverless function that checks S3 buckets for CSV files on a schedule
3. **ECS Fargate Container** - Processes CSV files found by Lambda (removes duplicates, cleans data)

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

### 4. Build & Push Docker Image (for ECS)
```bash
# Get ECR URL
ECR_URL=$(terraform output -raw ecr_repository_url)

# Authenticate & push
aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin $ECR_URL
docker build -t csv-handler .
docker tag csv-handler:latest $ECR_URL:latest
docker push $ECR_URL:latest
```

### 5. Test
```bash
# Upload a CSV file
aws s3 cp test.csv s3://<your-bucket>/test.csv

# Lambda will trigger ECS automatically, or manually invoke:
aws lambda invoke --function-name s3-file-checker response.json
```

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
    ↓ (triggers)
Lambda Function (checks S3 for CSV files)
    ↓ (triggers ECS task for each file)
ECS Fargate Container
    ↓ (downloads, processes, uploads)
S3 Bucket (stores cleaned CSV files)
```

## Project Structure

```
golang-csv-handler/
├── main.go              # CLI tool for local CSV processing
├── Dockerfile           # Docker image for ECS
├── entrypoint.sh        # Container entrypoint script
├── lambda/              # Lambda function code
│   ├── main.go         # Lambda handler (triggers ECS tasks)
│   └── Makefile        # Build automation
└── terraform/          # Infrastructure as Code
    ├── main.tf         # All AWS resources (Lambda, ECS, etc.)
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
- S3 bucket and all files
- Lambda function
- ECS cluster and task definitions
- ECR repository
- EventBridge schedule
- IAM roles and policies
- CloudWatch logs

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
- `terraform/terraform.tfvars` - Contains your actual deployment values
- `lambda/bootstrap` - Compiled binary (regenerate with `make build`)
- `terraform/lambda_function.zip` - Generated deployment package

Always use `terraform.tfvars.example` as a template and never commit your actual `terraform.tfvars` file.

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
