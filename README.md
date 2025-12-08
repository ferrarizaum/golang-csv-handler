# Golang CSV Handler

A Go-based project for handling CSV files with both a local CLI tool and an AWS Lambda function for automated S3 file checking.

## Project Overview

This project consists of two main components:

1. **CLI Tool** (`main.go`) - Command-line CSV handler for local file processing
2. **Lambda Function** (`lambda/`) - Serverless AWS Lambda function that checks S3 buckets for CSV files on a schedule

## Quick Start

See [QUICKSTART.md](./QUICKSTART.md) for a 5-minute setup guide.

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

## Project Structure

```
golang-csv-handler/
├── main.go              # CLI tool for local CSV processing
├── lambda/              # Lambda function code
│   ├── main.go         # Lambda handler
│   ├── Makefile        # Build automation
│   └── build.ps1       # Windows build script
├── terraform/          # Infrastructure as Code
│   ├── main.tf         # Resource definitions
│   ├── variables.tf    # Variable declarations
│   ├── outputs.tf      # Output values
│   └── terraform.tfvars.example  # Configuration template
├── DEPLOYMENT.md       # Detailed deployment guide
└── QUICKSTART.md       # Quick setup guide
```

## Documentation

- **[QUICKSTART.md](./QUICKSTART.md)** - Get started in 5 minutes
- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Comprehensive deployment guide
- **[lambda/README.md](./lambda/README.md)** - Lambda function documentation

## Cleanup

To remove all AWS resources:

```bash
cd terraform
terraform destroy
```

This will delete:
- S3 bucket and all files
- Lambda function
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
