# Terraform Infrastructure for CSV Handler

This directory contains the Terraform configuration for deploying the CSV handler infrastructure to AWS.

## Architecture

The infrastructure is organized using Terraform modules for better reusability and maintainability:

```
terraform/
├── main.tf              # Main orchestration file
├── variables.tf         # Input variables
├── outputs.tf          # Output values
├── locals.tf           # Local values and common tags
├── terraform.tfvars    # Variable values (customize this)
└── modules/
    ├── s3-bucket/           # S3 bucket with security settings
    ├── s3-folder/           # S3 folder/prefix creation
    ├── lambda-function/     # Lambda function with IAM and logging
    └── eventbridge-scheduler/ # EventBridge scheduler for Lambda
```

## Modules

### s3-bucket
Creates an S3 bucket with:
- Public access blocking
- Configurable lifecycle policies
- Tagging support

### s3-folder
Creates folder prefixes in S3 buckets (used with `for_each` for multiple folders).

### lambda-function
Creates a Lambda function with:
- IAM role and policies (using `aws_iam_policy_document`)
- CloudWatch log group
- S3 bucket permissions
- Environment variables

### eventbridge-scheduler
Creates an EventBridge Scheduler with:
- IAM role for invoking Lambda
- Configurable schedule expressions
- Enable/disable capability

## Best Practices Implemented

1. **DRY Principle**: No code repetition, everything is modularized
2. **Common Tags**: Centralized tagging using `locals.tf`
3. **IAM Policy Documents**: Using `data "aws_iam_policy_document"` instead of `jsonencode()`
4. **Module Outputs**: All modules expose useful outputs for composition
5. **Variable Validation**: Type constraints and descriptions for all variables
6. **Security**: Public access blocking enabled by default on S3

## Usage

### Initialize Terraform
```bash
cd terraform
terraform init
```

### Plan Changes
```bash
terraform plan
```

### Apply Changes
```bash
terraform apply
```

### Destroy Infrastructure
```bash
terraform destroy
```

## Customization

To customize the deployment, edit `terraform.tfvars`:

```hcl
aws_region           = "us-east-1"
s3_bucket_name       = "your-bucket-name"
lambda_function_name = "your-lambda-name"
schedule_expression  = "rate(5 minutes)"
environment          = "dev"
```

## Adding More Folders

To add more S3 folders, edit `locals.tf`:

```hcl
locals {
  s3_folders = ["input", "output", "archive", "temp", "backup"]
}
```

## Module Reusability

All modules can be reused in other projects. For example, to add another Lambda function:

```hcl
module "another_lambda" {
  source = "./modules/lambda-function"
  
  function_name = "another-function"
  # ... other parameters
}
```

## Outputs

After deployment, Terraform will display:
- S3 bucket name, ARN, and ID
- Lambda function name, ARN, and role ARN
- CloudWatch log group name and ARN
- EventBridge schedule name and ARN
- Scheduler role ARN
