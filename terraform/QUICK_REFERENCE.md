# Terraform Quick Reference

## 🚀 Quick Start

```bash
cd terraform
terraform init       # Install modules (first time only)
terraform plan       # Preview changes
terraform apply      # Deploy infrastructure
terraform destroy    # Remove all resources
```

## 📁 File Structure

```
terraform/
├── main.tf           ← Orchestrates all modules
├── locals.tf         ← Common tags & shared values
├── variables.tf      ← Input variables
├── outputs.tf        ← Output values
├── terraform.tfvars  ← Your custom values
└── modules/          ← Reusable components
    ├── s3-bucket/
    ├── s3-folder/
    ├── lambda-function/
    └── eventbridge-scheduler/
```

## 🔧 Common Tasks

### Add More S3 Folders
Edit `locals.tf`:
```hcl
locals {
  s3_folders = ["input", "output", "archive", "temp"]  # Add here
}
```

### Change Schedule
Edit `terraform.tfvars`:
```hcl
schedule_expression = "rate(10 minutes)"  # or "cron(0 9 * * ? *)"
```

### Change Lambda Timeout
Edit `main.tf`, find `module "csv_handler_lambda"`:
```hcl
timeout = 60  # seconds
```

### Change Log Retention
Edit `main.tf`, find `module "csv_handler_lambda"`:
```hcl
log_retention_days = 14  # days
```

### Add Tags
Edit `locals.tf`:
```hcl
locals {
  common_tags = {
    Environment = var.environment
    Project     = "csv-handler"
    ManagedBy   = "terraform"
    Team        = "data-engineering"  # Add custom tags
    CostCenter  = "engineering"
  }
}
```

## 📊 View Outputs

After `terraform apply`, see all outputs:
```bash
terraform output
```

Get specific output:
```bash
terraform output s3_bucket_name
terraform output lambda_function_arn
```

## 🔍 Debugging

### View State
```bash
terraform show
terraform state list
```

### View Specific Resource
```bash
terraform state show module.csv_bucket.aws_s3_bucket.this
```

### Format Code
```bash
terraform fmt -recursive
```

### Validate Configuration
```bash
terraform validate
```

## 🎯 Module Usage

### Using s3-bucket Module
```hcl
module "my_bucket" {
  source = "./modules/s3-bucket"
  
  bucket_name         = "my-unique-bucket-name"
  prevent_destroy     = false
  block_public_access = true
  tags                = local.common_tags
}
```

### Using lambda-function Module
```hcl
module "my_lambda" {
  source = "./modules/lambda-function"
  
  function_name    = "my-function"
  filename         = "my-function.zip"
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("my-function.zip")
  timeout          = 30
  
  environment_variables = {
    MY_VAR = "my-value"
  }
  
  s3_bucket_arn = module.my_bucket.bucket_arn
  tags          = local.common_tags
}
```

### Using eventbridge-scheduler Module
```hcl
module "my_scheduler" {
  source = "./modules/eventbridge-scheduler"
  
  schedule_name       = "my-schedule"
  schedule_expression = "rate(5 minutes)"
  lambda_function_arn = module.my_lambda.function_arn
  enabled             = true
  tags                = local.common_tags
}
```

## 🔐 Security Checklist

- ✅ S3 public access blocked
- ✅ IAM least privilege principle
- ✅ CloudWatch logging enabled
- ✅ No hardcoded credentials
- ✅ Tags for cost tracking

## 📝 Variable Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `aws_region` | AWS region | `us-east-1` |
| `s3_bucket_name` | S3 bucket name | `my-csv-bucket` |
| `lambda_function_name` | Lambda name | `csv-handler` |
| `schedule_expression` | Cron/rate | `rate(5 minutes)` |
| `environment` | Environment | `dev`, `staging`, `prod` |

## 🎨 Schedule Expression Examples

```hcl
# Rate-based
schedule_expression = "rate(1 minute)"
schedule_expression = "rate(5 minutes)"
schedule_expression = "rate(1 hour)"
schedule_expression = "rate(1 day)"

# Cron-based (UTC timezone)
schedule_expression = "cron(0 9 * * ? *)"      # Every day at 9:00 AM
schedule_expression = "cron(0 */2 * * ? *)"    # Every 2 hours
schedule_expression = "cron(0 9 ? * MON-FRI *)" # Weekdays at 9:00 AM
schedule_expression = "cron(0 0 1 * ? *)"      # First day of month
```

## 🚨 Troubleshooting

### Module Not Found
```bash
terraform init  # Reinstall modules
```

### Resource Already Exists
```bash
terraform import <resource_address> <resource_id>
```

### State Lock Error
```bash
# If using S3 backend with DynamoDB
# Manually remove lock from DynamoDB table
```

### Plan Shows Unwanted Changes
```bash
terraform refresh  # Sync state with reality
terraform plan     # Check again
```

## 📚 Learn More

- [Terraform Documentation](https://www.terraform.io/docs)
- [AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Module Best Practices](https://www.terraform.io/docs/language/modules/develop/index.html)

## 💡 Pro Tips

1. **Always run `terraform plan` before `apply`**
2. **Use workspaces for multiple environments**
3. **Store state remotely (S3 + DynamoDB)**
4. **Version your modules**
5. **Use `.terraform.lock.hcl` in version control**
6. **Never commit `terraform.tfvars` with secrets**
7. **Use `terraform fmt` before committing**
8. **Tag everything for cost tracking**

## 🔄 State Migration Commands

If you need to move resources from old to new structure:

```bash
# S3 Bucket
terraform state mv aws_s3_bucket.csv_bucket module.csv_bucket.aws_s3_bucket.this
terraform state mv aws_s3_bucket_public_access_block.csv_bucket module.csv_bucket.aws_s3_bucket_public_access_block.this

# S3 Folders
terraform state mv 'aws_s3_object.input' 'module.s3_folders["input"].aws_s3_object.folder'
terraform state mv 'aws_s3_object.output' 'module.s3_folders["output"].aws_s3_object.folder'
terraform state mv 'aws_s3_object.archive' 'module.s3_folders["archive"].aws_s3_object.folder'

# Lambda
terraform state mv aws_iam_role.lambda_role module.csv_handler_lambda.aws_iam_role.lambda
terraform state mv aws_iam_role_policy.lambda_policy module.csv_handler_lambda.aws_iam_role_policy.lambda
terraform state mv aws_cloudwatch_log_group.lambda_logs module.csv_handler_lambda.aws_cloudwatch_log_group.lambda
terraform state mv aws_lambda_function.s3_checker module.csv_handler_lambda.aws_lambda_function.this

# Scheduler
terraform state mv aws_iam_role.scheduler_role module.lambda_scheduler.aws_iam_role.scheduler
terraform state mv aws_iam_role_policy.scheduler_policy module.lambda_scheduler.aws_iam_role_policy.scheduler
terraform state mv aws_scheduler_schedule.lambda_schedule module.lambda_scheduler.aws_scheduler_schedule.this
```

## 📞 Need Help?

Check these files:
- `README.md` - Detailed usage guide
- `REFACTORING_SUMMARY.md` - What changed and why
- `ARCHITECTURE.md` - System architecture diagrams
