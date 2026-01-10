# Outputs in Terraform allow you to see important information after deployment
# These values are displayed after running terraform apply

# Output the S3 bucket name
output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.csv_bucket.id
}

# Output the S3 bucket ARN (Amazon Resource Name)
output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.csv_bucket.arn
}

# Output the Lambda function name
output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.s3_checker.function_name
}

# Output the Lambda function ARN
output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.s3_checker.arn
}

# Output the EventBridge schedule name
output "schedule_name" {
  description = "Name of the EventBridge schedule"
  value       = aws_scheduler_schedule.lambda_schedule.name
}

# Output the CloudWatch log group name
output "log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}