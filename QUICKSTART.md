# Quick Start Guide

Get up and running in 5 minutes!

## Prerequisites Checklist

- [ ] AWS Account created
- [ ] AWS CLI installed and configured (`aws configure`)
- [ ] Terraform installed (`terraform version`)
- [ ] Go installed (`go version`)

## Deployment Steps

### 1. Build Lambda (30 seconds)

```bash
# Windows
cd lambda
make build
# Or: ..\build.bat

# Linux/Mac
cd lambda
make build
# Or: ../build.sh
```

### 2. Configure Terraform (1 minute)

**Important:** The `terraform.tfvars` file is not in git (for security). You need to create it from the example template.

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` - **change the bucket name** (must be globally unique):
```hcl
s3_bucket_name = "my-csv-bucket-12345"  # Change this! Use random numbers/letters
```

**Note:** S3 bucket names can only contain lowercase letters, numbers, dots, and hyphens (no underscores!).

### 3. Deploy (2 minutes)

```bash
terraform init
terraform plan   # Review what will be created
terraform apply  # Type 'yes' to confirm
```

### 4. Test (1 minute)

```bash
# Upload a test file
echo "test,data" > test.csv
aws s3 cp test.csv s3://<your-bucket-name>/test.csv

# Manually trigger Lambda
aws lambda invoke --function-name s3-file-checker --payload '{}' response.json
cat response.json
```

### 5. View Logs

```bash
aws logs tail /aws/lambda/s3-file-checker --follow
```

## What You Get

✅ S3 bucket for CSV files  
✅ Lambda function that checks S3  
✅ EventBridge schedule (runs every hour by default)  
✅ CloudWatch logs for monitoring  

## Next Steps

- Upload CSV files to S3
- Lambda will check automatically on schedule
- View results in CloudWatch Logs
- Extend Lambda to process CSV files

## Troubleshooting

**Error: Bucket name exists**
→ Change `s3_bucket_name` in `terraform.tfvars` to something more unique

**Error: Lambda build failed**
→ Make sure Go is installed: `go version`

**Error: Terraform not found**
→ Install Terraform: https://www.terraform.io/downloads

## Full Documentation

- **[DEPLOYMENT.md](DEPLOYMENT.md)**: Detailed deployment guide
- **[terraform/README.md](terraform/README.md)**: Terraform documentation
- **[lambda/README.md](lambda/README.md)**: Lambda function docs

