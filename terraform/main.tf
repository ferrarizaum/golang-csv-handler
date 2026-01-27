# Main Terraform configuration file
# This file orchestrates all the modules to create the CSV handler infrastructure

# Configure the AWS Provider
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
provider "aws" {
  region = var.aws_region
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Create a ZIP file containing the Lambda function binary
data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "../lambda/bootstrap"
  output_path = "lambda_function.zip"
}

# ============================================================================
# S3 Bucket Module
# ============================================================================
module "csv_bucket" {
  source = "./modules/s3-bucket"

  bucket_name         = var.s3_bucket_name
  prevent_destroy     = false
  block_public_access = true
  tags                = local.common_tags
}

# ============================================================================
# S3 Folders Module
# ============================================================================
module "s3_folders" {
  source   = "./modules/s3-folder"
  for_each = toset(local.s3_folders)

  bucket_id   = module.csv_bucket.bucket_id
  folder_name = each.value
}

# ============================================================================
# Lambda Function Module
# ============================================================================
module "csv_handler_lambda" {
  source = "./modules/lambda-function"

  function_name    = var.lambda_function_name
  filename         = data.archive_file.lambda_zip.output_path
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  architectures    = ["x86_64"]
  timeout          = 30

  environment_variables = {
    S3_BUCKET_NAME = var.s3_bucket_name
  }

  s3_bucket_arn      = module.csv_bucket.bucket_arn
  log_retention_days = 7

  tags = local.common_tags
}

# ============================================================================
# EventBridge Scheduler Module
# ============================================================================
module "lambda_scheduler" {
  source = "./modules/eventbridge-scheduler"

  schedule_name       = "${var.lambda_function_name}-schedule"
  schedule_expression = var.schedule_expression
  lambda_function_arn = module.csv_handler_lambda.function_arn
  enabled             = true

  tags = local.common_tags
}