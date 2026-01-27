# Variables for the EventBridge Scheduler module

variable "schedule_name" {
  description = "Name of the EventBridge schedule"
  type        = string
}

variable "schedule_group_name" {
  description = "Name of the EventBridge schedule group"
  type        = string
  default     = "default"
}

variable "schedule_expression" {
  description = "Schedule expression (rate or cron format)"
  type        = string
}

variable "lambda_function_arn" {
  description = "ARN of the Lambda function to invoke"
  type        = string
}

variable "flexible_time_window_mode" {
  description = "Flexible time window mode (OFF or FLEXIBLE)"
  type        = string
  default     = "OFF"
}

variable "input_json" {
  description = "Optional JSON input to pass to the Lambda function"
  type        = string
  default     = null
}

variable "enabled" {
  description = "Whether the schedule is enabled"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
