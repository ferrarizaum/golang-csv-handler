# Variables in Terraform allow you to customize your infrastructure
# without changing the main configuration files.
# You can set these in terraform.tfvars, environment variables, or command line.

# AWS region where resources will be created
# Region is the geographic location of AWS data centers
variable "aws_region" {
  description = "AWS region where resources will be created"
  type        = string
  default     = "us-east-1" # Default to US East (N. Virginia)
}

# Environment name (dev, staging, prod, etc.)
# Useful for organizing resources and managing multiple environments
variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

# Name of the S3 bucket
# S3 bucket names must be globally unique across all AWS accounts
variable "s3_bucket_name" {
  description = "Name of the S3 bucket to check for files (must be globally unique)"
  type        = string
  # No default - you must provide this value
  # Example: "my-csv-files-bucket-12345"
}

# Lambda function name
variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "s3-file-checker"
}

# EventBridge schedule expression
# Defines when the Lambda should be triggered
# Examples:
#   - "rate(5 minutes)" - Every 5 minutes
#   - "rate(1 hour)" - Every hour
#   - "cron(0 9 * * ? *)" - Every day at 9:00 AM UTC
#   - "cron(0 */6 * * ? *)" - Every 6 hours
variable "schedule_expression" {
  description = "EventBridge schedule expression (cron or rate format)"
  type        = string
  default     = "rate(1 hour)" # Default: Check every hour
}

# ECS Service name
# This will be used to name ECS resources (cluster, task definition, etc.)
variable "ecs_service_name" {
  description = "Name of the ECS service and related resources"
  type        = string
  default     = "csv-handler"
}

# ECS Task CPU
# Fargate CPU is specified in CPU units (1024 = 1 vCPU)
# Valid values: 256, 512, 1024, 2048, 4096
# 256 = 0.25 vCPU (good for small tasks)
# 512 = 0.5 vCPU
# 1024 = 1 vCPU (good for medium tasks)
variable "ecs_task_cpu" {
  description = "CPU units for ECS Fargate task (256 = 0.25 vCPU, 512 = 0.5 vCPU, 1024 = 1 vCPU)"
  type        = number
  default     = 512 # 0.5 vCPU - good balance for CSV processing
}

# ECS Task Memory
# Memory in MB. Must be compatible with CPU:
# - 256 CPU: 512-2048 MB
# - 512 CPU: 1024-4096 MB
# - 1024 CPU: 2048-8192 MB
# - 2048 CPU: 4096-16384 MB
variable "ecs_task_memory" {
  description = "Memory in MB for ECS Fargate task"
  type        = number
  default     = 1024 # 1 GB - sufficient for most CSV files
}

