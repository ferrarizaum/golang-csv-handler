# EventBridge Scheduler Module
# This module creates an EventBridge Scheduler to trigger a Lambda function

# IAM policy document for EventBridge Scheduler assume role
data "aws_iam_policy_document" "scheduler_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

# IAM role for EventBridge Scheduler
resource "aws_iam_role" "scheduler" {
  name               = "${var.schedule_name}-role"
  assume_role_policy = data.aws_iam_policy_document.scheduler_assume_role.json

  tags = var.tags
}

# IAM policy document for invoking Lambda
data "aws_iam_policy_document" "scheduler_invoke_lambda" {
  statement {
    effect    = "Allow"
    actions   = ["lambda:InvokeFunction"]
    resources = [var.lambda_function_arn]
  }
}

# Attach policy to Scheduler role
resource "aws_iam_role_policy" "scheduler" {
  name   = "${var.schedule_name}-policy"
  role   = aws_iam_role.scheduler.id
  policy = data.aws_iam_policy_document.scheduler_invoke_lambda.json
}

# EventBridge Scheduler
resource "aws_scheduler_schedule" "this" {
  name       = var.schedule_name
  group_name = var.schedule_group_name

  schedule_expression = var.schedule_expression

  flexible_time_window {
    mode = var.flexible_time_window_mode
  }

  target {
    arn      = var.lambda_function_arn
    role_arn = aws_iam_role.scheduler.arn

    # Optional: Pass custom input to Lambda
    input = var.input_json
  }

  state = var.enabled ? "ENABLED" : "DISABLED"
}
