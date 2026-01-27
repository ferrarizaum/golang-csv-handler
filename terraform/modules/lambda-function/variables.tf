# Variables for the Lambda function module

variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "filename" {
  description = "Path to the Lambda deployment package"
  type        = string
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
}

variable "runtime" {
  description = "Lambda runtime"
  type        = string
}

variable "source_code_hash" {
  description = "Hash of the Lambda deployment package"
  type        = string
}

variable "architectures" {
  description = "Instruction set architecture for the Lambda function"
  type        = list(string)
  default     = ["x86_64"]
}

variable "timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30
}

variable "environment_variables" {
  description = "Environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "s3_bucket_arn" {
  description = "ARN of S3 bucket to grant access to (optional)"
  type        = string
  default     = null
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 7
}

variable "additional_policy_statements" {
  description = "Additional IAM policy statements for the Lambda function"
  type = list(object({
    effect    = string
    actions   = list(string)
    resources = list(string)
  }))
  default = []
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
