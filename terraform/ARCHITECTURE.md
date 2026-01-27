# Terraform Architecture

## Module Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                          main.tf                                 │
│                   (Orchestration Layer)                          │
└────────────┬────────────┬────────────┬─────────────┬────────────┘
             │            │            │             │
             ▼            ▼            ▼             ▼
    ┌────────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
    │ s3-bucket  │ │s3-folder │ │  lambda  │ │ eventbridge  │
    │   module   │ │  module  │ │  module  │ │   module     │
    └────────────┘ └──────────┘ └──────────┘ └──────────────┘
         │              │            │              │
         │              │            │              │
         ▼              ▼            ▼              ▼
    ┌─────────────────────────────────────────────────────┐
    │                  AWS Resources                       │
    │  • S3 Bucket           • Lambda Function            │
    │  • Public Access Block • IAM Role & Policy          │
    │  • S3 Folders          • CloudWatch Log Group       │
    │                        • EventBridge Schedule       │
    └─────────────────────────────────────────────────────┘
```

## Data Flow

```
┌──────────────────────────────────────────────────────────────┐
│  1. User uploads CSV to S3 input/ folder                     │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────┐
│  2. EventBridge Scheduler triggers Lambda (every 5 minutes)  │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────┐
│  3. Lambda checks S3 input/ folder                           │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────┐
│  4. Lambda processes CSV files                               │
│     • Reads from input/                                      │
│     • Processes data                                         │
│     • Writes to output/                                      │
│     • Moves original to archive/                             │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────┐
│  5. Logs written to CloudWatch                               │
└──────────────────────────────────────────────────────────────┘
```

## Module Dependencies

```
locals.tf (common_tags)
    │
    ├──> module.csv_bucket
    │        │
    │        └──> module.s3_folders (depends on bucket_id)
    │
    ├──> module.csv_handler_lambda (depends on bucket_arn)
    │        │
    │        └──> module.lambda_scheduler (depends on function_arn)
    │
    └──> All modules receive common_tags
```

## File Organization

```
terraform/
│
├── Configuration Files
│   ├── main.tf                    # Orchestration (95 lines)
│   ├── variables.tf               # Input variables
│   ├── outputs.tf                 # Output values
│   ├── locals.tf                  # Common values & tags
│   └── terraform.tfvars           # Variable values
│
├── Documentation
│   ├── README.md                  # Usage guide
│   ├── REFACTORING_SUMMARY.md     # What changed
│   └── ARCHITECTURE.md            # This file
│
└── modules/                       # Reusable modules
    │
    ├── s3-bucket/
    │   ├── main.tf                # S3 + public access block
    │   ├── variables.tf           # Module inputs
    │   └── outputs.tf             # bucket_id, bucket_arn
    │
    ├── s3-folder/
    │   ├── main.tf                # S3 object (folder)
    │   ├── variables.tf           # bucket_id, folder_name
    │   └── outputs.tf             # folder_key, folder_id
    │
    ├── lambda-function/
    │   ├── main.tf                # Lambda + IAM + CloudWatch
    │   ├── variables.tf           # Function configuration
    │   └── outputs.tf             # function_arn, role_arn
    │
    └── eventbridge-scheduler/
        ├── main.tf                # Scheduler + IAM
        ├── variables.tf           # Schedule configuration
        └── outputs.tf             # schedule_arn, role_arn
```

## Resource Relationships

```
aws_s3_bucket (csv_bucket)
    │
    ├──> aws_s3_bucket_public_access_block
    │
    └──> aws_s3_object (input/, output/, archive/)

aws_iam_role (lambda_role)
    │
    ├──> aws_iam_role_policy (lambda_policy)
    │        │
    │        └──> Permissions:
    │             • S3: ListBucket, GetObject, PutObject, DeleteObject
    │             • CloudWatch: CreateLogGroup, CreateLogStream, PutLogEvents
    │
    └──> aws_lambda_function (s3_checker)
             │
             └──> aws_cloudwatch_log_group (lambda_logs)

aws_iam_role (scheduler_role)
    │
    ├──> aws_iam_role_policy (scheduler_policy)
    │        │
    │        └──> Permissions:
    │             • Lambda: InvokeFunction
    │
    └──> aws_scheduler_schedule (lambda_schedule)
             │
             └──> Triggers: aws_lambda_function.s3_checker
```

## Security Model

```
┌─────────────────────────────────────────────────────────────┐
│                      S3 Bucket                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Public Access Block (ENABLED)                        │  │
│  │  • block_public_acls = true                           │  │
│  │  • block_public_policy = true                         │  │
│  │  • ignore_public_acls = true                          │  │
│  │  • restrict_public_buckets = true                     │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             │
                             │ Access granted via IAM
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    Lambda IAM Role                           │
│  • Principle of Least Privilege                             │
│  • Only necessary S3 permissions                            │
│  • CloudWatch logging only                                  │
│  • No wildcard permissions                                  │
└─────────────────────────────────────────────────────────────┘
```

## Tagging Strategy

All resources are tagged with:

```hcl
{
  Environment = "dev" | "staging" | "prod"
  Project     = "csv-handler"
  ManagedBy   = "terraform"
  Name        = "<resource-specific-name>"
}
```

This enables:
- Cost tracking by environment
- Resource grouping by project
- Identification of IaC-managed resources
- Easy filtering in AWS Console

## Scalability

### Adding More Lambda Functions

```hcl
module "another_lambda" {
  source = "./modules/lambda-function"
  
  function_name = "another-function"
  filename      = "another-function.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  # ... other parameters
  
  tags = local.common_tags
}
```

### Adding More S3 Buckets

```hcl
module "another_bucket" {
  source = "./modules/s3-bucket"
  
  bucket_name = "another-bucket-name"
  tags        = local.common_tags
}
```

### Adding More Folders

Edit `locals.tf`:
```hcl
locals {
  s3_folders = ["input", "output", "archive", "temp", "backup"]
}
```

## Cost Optimization

- **Lambda**: Timeout set to 30s (adjust based on actual needs)
- **CloudWatch Logs**: 7-day retention (adjust based on compliance needs)
- **S3**: Standard storage class (consider lifecycle policies for archiving)
- **EventBridge**: Minimal cost for schedule-based triggers

## Monitoring

All resources emit logs/metrics to CloudWatch:

```
CloudWatch Logs
    └── /aws/lambda/csv-handler
            ├── Execution logs
            ├── Error traces
            └── Custom application logs

CloudWatch Metrics
    ├── Lambda Invocations
    ├── Lambda Duration
    ├── Lambda Errors
    └── S3 Bucket Metrics
```

## Disaster Recovery

To backup/restore the infrastructure:

1. **State File**: Store in S3 with versioning enabled
2. **Code**: Version control (Git)
3. **Modules**: Tag versions for stability
4. **Data**: Enable S3 versioning for CSV files

## Next Steps

1. Run `terraform init` to install modules
2. Run `terraform plan` to preview changes
3. Run `terraform apply` to deploy
4. Monitor CloudWatch logs for Lambda execution
5. Test by uploading a CSV file to `input/` folder
