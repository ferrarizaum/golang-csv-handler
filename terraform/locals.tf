# Local values for reuse across the configuration
# This helps maintain consistency and reduces repetition

locals {
  # Common tags applied to all resources
  common_tags = {
    Environment = var.environment
    Project     = "csv-handler"
    ManagedBy   = "terraform"
  }

  # S3 folders to create
  s3_folders = ["input", "output", "archive"]
}
