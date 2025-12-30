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
      ECS_CLUSTER_NAME      = aws_ecs_cluster.csv_handler.name
      ECS_TASK_DEFINITION   = aws_ecs_task_definition.csv_handler.family
      ECS_SUBNET_IDS        = join(",", data.aws_subnets.default.ids)
      ECS_SECURITY_GROUP_ID = aws_security_group.ecs_task.id
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

# ============================================================================
# ECS FARGATE INFRASTRUCTURE
# ============================================================================
# ECS Fargate is a serverless container platform that runs your containers
# without you needing to manage servers. Lambda will trigger ECS tasks to
# process CSV files found in S3.

# Create an ECR (Elastic Container Registry) repository
# ECR is AWS's Docker registry - like Docker Hub, but private and integrated with AWS
# This is where we'll store the Docker image of our Go application
resource "aws_ecr_repository" "csv_handler" {
  name                 = "${var.ecs_service_name}-repo"
  image_tag_mutability = "MUTABLE" # Allows overwriting tags (good for dev)

  # Configure image scanning for security vulnerabilities
  image_scanning_configuration {
    scan_on_push = true # Automatically scan images when pushed
  }

  tags = {
    Name        = "${var.ecs_service_name}-repo"
    Environment = var.environment
  }
}

# Set lifecycle policy for ECR to automatically delete old images
# This prevents the repository from growing too large and saves storage costs
resource "aws_ecr_lifecycle_policy" "csv_handler" {
  repository = aws_ecr_repository.csv_handler.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# Get the default VPC (simplest option for beginners)
# VPC (Virtual Private Cloud) is your isolated network in AWS
# The default VPC comes pre-configured with internet access
data "aws_vpc" "default" {
  default = true
}

# Get the default subnets from the default VPC
# Subnets are like separate network segments within your VPC
# We need at least 2 subnets in different availability zones for ECS
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Create a security group for ECS tasks
# Security groups are like firewalls - they control what traffic can enter/exit
# For Fargate, we mainly need outbound access to S3 and ECR
resource "aws_security_group" "ecs_task" {
  name        = "${var.ecs_service_name}-task-sg"
  description = "Security group for ECS Fargate tasks"
  vpc_id      = data.aws_vpc.default.id

  # Allow all outbound traffic (needed for S3, ECR, CloudWatch)
  # In production, you might want to restrict this more
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1" # All protocols
    cidr_blocks = ["0.0.0.0/0"] # All IPs (needed for internet access)
  }

  tags = {
    Name        = "${var.ecs_service_name}-task-sg"
    Environment = var.environment
  }
}

# Create an IAM role for ECS tasks
# This role defines what permissions the containers have (e.g., access to S3)
resource "aws_iam_role" "ecs_task_role" {
  name = "${var.ecs_service_name}-task-role"

  # Trust policy - allows ECS to assume this role
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.ecs_service_name}-task-role"
    Environment = var.environment
  }
}

# IAM policy for ECS tasks to access S3
# This allows the container to read input files and write output files
resource "aws_iam_role_policy" "ecs_task_s3_policy" {
  name = "${var.ecs_service_name}-task-s3-policy"
  role = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
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

# Create an IAM role for ECS task execution
# This is different from the task role - it's used by ECS itself to pull images, etc.
resource "aws_iam_role" "ecs_execution_role" {
  name = "${var.ecs_service_name}-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.ecs_service_name}-execution-role"
    Environment = var.environment
  }
}

# Attach the AWS managed policy for ECS task execution
# This policy allows ECS to pull images from ECR and write logs to CloudWatch
resource "aws_iam_role_policy_attachment" "ecs_execution_role_policy" {
  role       = aws_iam_role.ecs_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Create an ECS Cluster
# A cluster is a logical grouping of tasks and services
resource "aws_ecs_cluster" "csv_handler" {
  name = "${var.ecs_service_name}-cluster"

  # Enable CloudWatch Container Insights for monitoring
  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name        = "${var.ecs_service_name}-cluster"
    Environment = var.environment
  }
}

# Create an ECS Task Definition
# This defines how to run your container: which image, CPU, memory, environment variables, etc.
resource "aws_ecs_task_definition" "csv_handler" {
  family                   = var.ecs_service_name
  network_mode             = "awsvpc" # Required for Fargate
  requires_compatibilities = ["FARGATE"] # Use Fargate (serverless)
  cpu                      = var.ecs_task_cpu # CPU units (256 = 0.25 vCPU)
  memory                   = var.ecs_task_memory # Memory in MB

  # IAM roles
  execution_role_arn = aws_iam_role.ecs_execution_role.arn
  task_role_arn     = aws_iam_role.ecs_task_role.arn

  # Container definition
  container_definitions = jsonencode([
    {
      name  = "csv-handler"
      image = "${aws_ecr_repository.csv_handler.repository_url}:latest"

      # Logging configuration - sends logs to CloudWatch
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs_logs.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "ecs"
        }
      }

      # Environment variables passed to the container
      # These will be set by Lambda when it runs the task
      environment = [
        {
          name  = "S3_BUCKET_NAME"
          value = var.s3_bucket_name
        }
      ]

      # The container runs the CSV handler with input/output paths
      # These will be passed as command overrides when Lambda runs the task
      # For now, we set defaults (they'll be overridden)
      command = [
        "-input", "/tmp/input.csv",
        "-output", "/tmp/output.csv"
      ]

      # Essential means if this container stops, the task stops
      essential = true
    }
  ])

  tags = {
    Name        = var.ecs_service_name
    Environment = var.environment
  }
}

# Create CloudWatch Log Group for ECS logs
resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/${var.ecs_service_name}"
  retention_in_days = 7

  tags = {
    Name        = "${var.ecs_service_name}-logs"
    Environment = var.environment
  }
}

# Update Lambda IAM role to allow invoking ECS RunTask
# This allows Lambda to trigger ECS tasks when it finds files
resource "aws_iam_role_policy" "lambda_ecs_policy" {
  name = "${var.lambda_function_name}-ecs-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:RunTask"
        ]
        Resource = aws_ecs_task_definition.csv_handler.arn
      },
      {
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = [
          aws_iam_role.ecs_task_role.arn,
          aws_iam_role.ecs_execution_role.arn
        ]
      }
    ]
  })
}
