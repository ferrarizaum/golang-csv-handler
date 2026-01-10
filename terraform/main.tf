# Main Terraform configuration file
# This file sets up the AWS provider and defines all the resources

# Configure the AWS Provider
# This tells Terraform to use AWS and which region to use
terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0"
    }
  }
}

# Configure the AWS Provider
# This uses the region from the variable
provider "aws" {
  region = var.aws_region
}

# Data source to get the current AWS account ID and region
# This is useful for creating IAM policies
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Create an S3 bucket for storing CSV files
# S3 is AWS's object storage service (like a file system in the cloud)
resource "aws_s3_bucket" "csv_bucket" {
  bucket = var.s3_bucket_name

  # Prevent accidental deletion
  lifecycle {
    prevent_destroy = false
  }

  tags = {
    Name        = var.s3_bucket_name
    Environment = var.environment
  }
}

# Create a folder in the S3 bucket
resource "aws_s3_object" "csv_folder" {
  bucket = aws_s3_bucket.csv_bucket.id
  key    = "csv-files/"
}

# Block public access to the S3 bucket
# This is a security best practice - only allow access from authorized sources
resource "aws_s3_bucket_public_access_block" "csv_bucket" {
  bucket = aws_s3_bucket.csv_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls       = true
  restrict_public_buckets  = true
}

# Create an IAM role for the Lambda function
# IAM roles define what permissions the Lambda function has
resource "aws_iam_role" "lambda_role" {
  name = "${var.lambda_function_name}-role"

  # Trust policy - allows Lambda service to assume this role
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.lambda_function_name}-role"
    Environment = var.environment
  }
}

# IAM policy that allows Lambda to:
# - List objects in the S3 bucket
# - Get object metadata from S3
# - Write logs to CloudWatch
resource "aws_iam_role_policy" "lambda_policy" {
  name = "${var.lambda_function_name}-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObject"
        ]
        Resource = [
          aws_s3_bucket.csv_bucket.arn,
          "${aws_s3_bucket.csv_bucket.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      }
    ]
  })
}

# Create a ZIP file containing the Lambda function binary
# Lambda requires code to be packaged as a ZIP file
data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "../lambda/bootstrap"
  output_path = "lambda_function.zip"
}

# Create the Lambda function
# This is the serverless function that will check S3 for files
resource "aws_lambda_function" "s3_checker" {
  filename         = data.archive_file.lambda_zip.output_path
  function_name    = var.lambda_function_name
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  runtime          = "provided.al2023" # Go runtime for Lambda
  architectures   = ["x86_64"] # Match the amd64 build target

  # Environment variables passed to the Lambda function
  environment {
    variables = {
      S3_BUCKET_NAME        = var.s3_bucket_name
    }
  }

  # Timeout in seconds (how long Lambda can run before being terminated)
  timeout = 30

  # Reserved concurrent executions limits how many instances can run simultaneously
  # Note: Removed due to account concurrency limits. For cost control, monitor usage via CloudWatch.
  # If your account has higher limits, you can uncomment and set to 1:
  # reserved_concurrent_executions = 1

  tags = {
    Name        = var.lambda_function_name
    Environment = var.environment
  }
}

# Create a CloudWatch Log Group for Lambda logs
# CloudWatch is AWS's logging and monitoring service
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 7 # Keep logs for 7 days

  tags = {
    Name        = "${var.lambda_function_name}-logs"
    Environment = var.environment
  }
}

# Create an IAM role for EventBridge Scheduler
# EventBridge needs permission to invoke the Lambda function
resource "aws_iam_role" "scheduler_role" {
  name = "${var.lambda_function_name}-scheduler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "scheduler.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.lambda_function_name}-scheduler-role"
    Environment = var.environment
  }
}

# IAM policy that allows EventBridge to invoke the Lambda function
resource "aws_iam_role_policy" "scheduler_policy" {
  name = "${var.lambda_function_name}-scheduler-policy"
  role = aws_iam_role.scheduler_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction"
        ]
        Resource = aws_lambda_function.s3_checker.arn
      }
    ]
  })
}

# Create EventBridge Scheduler to trigger Lambda on a schedule
# EventBridge Scheduler is AWS's cron service
resource "aws_scheduler_schedule" "lambda_schedule" {
  name       = "${var.lambda_function_name}-schedule"
  group_name = "default"

  # Schedule expression (cron or rate format)
  # Examples:
  #   - "rate(5 minutes)" - Every 5 minutes
  #   - "rate(1 hour)" - Every hour
  #   - "cron(0 9 * * ? *)" - Every day at 9:00 AM UTC
  schedule_expression = var.schedule_expression

  # Flexible time window (Lambda can be invoked within this window)
  flexible_time_window {
    mode = "OFF"
  }

  # Target - what to invoke when the schedule triggers
  target {
    arn      = aws_lambda_function.s3_checker.arn
    role_arn = aws_iam_role.scheduler_role.arn
  }

  state = "ENABLED" # Enable the schedule
}