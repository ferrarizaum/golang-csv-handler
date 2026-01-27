# Outputs in Terraform allow you to see important information after deployment
# These values are displayed after running terraform apply

# ============================================================================
# S3 Bucket Outputs
# ============================================================================
output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = module.csv_bucket.bucket_name
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = module.csv_bucket.bucket_arn
}

output "s3_bucket_id" {
  description = "ID of the S3 bucket"
  value       = module.csv_bucket.bucket_id
}

# ============================================================================
# Lambda Function Outputs
# ============================================================================
output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = module.csv_handler_lambda.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = module.csv_handler_lambda.function_arn
}

output "lambda_role_arn" {
  description = "ARN of the Lambda IAM role"
  value       = module.csv_handler_lambda.role_arn
}

output "log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = module.csv_handler_lambda.log_group_name
}

output "log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = module.csv_handler_lambda.log_group_arn
}

# ============================================================================
# EventBridge Scheduler Outputs
# ============================================================================
output "schedule_name" {
  description = "Name of the EventBridge schedule"
  value       = module.lambda_scheduler.schedule_name
}

output "schedule_arn" {
  description = "ARN of the EventBridge schedule"
  value       = module.lambda_scheduler.schedule_arn
}

output "scheduler_role_arn" {
  description = "ARN of the scheduler IAM role"
  value       = module.lambda_scheduler.scheduler_role_arn
}