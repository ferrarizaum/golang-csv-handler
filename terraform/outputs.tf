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

# ============================================================================
# ECS FARGATE OUTPUTS
# ============================================================================

# Output the ECR repository URL
# Use this to push your Docker image: docker push <this-url>:latest
output "ecr_repository_url" {
  description = "URL of the ECR repository (use this to push Docker images)"
  value       = aws_ecr_repository.csv_handler.repository_url
}

# Output the ECS cluster name
output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.csv_handler.name
}

# Output the ECS task definition ARN
# Lambda will use this to run tasks
output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.csv_handler.arn
}

# Output the ECS task definition family
output "ecs_task_definition_family" {
  description = "Family name of the ECS task definition"
  value       = aws_ecs_task_definition.csv_handler.family
}

# Output the subnet IDs (needed for running tasks)
output "ecs_subnet_ids" {
  description = "Subnet IDs where ECS tasks will run"
  value       = data.aws_subnets.default.ids
}

# Output the security group ID for ECS tasks
output "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  value       = aws_security_group.ecs_task.id
}
