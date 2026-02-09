# S3 Bucket Module
# This module creates an S3 bucket with security best practices

resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name

  # lifecycle {
  #   prevent_destroy = var.prevent_destroy
  # }

  tags = merge(
    var.tags,
    {
      Name = var.bucket_name
    }
  )
}

# Block all public access to the bucket
resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = var.block_public_access
  block_public_policy     = var.block_public_access
  ignore_public_acls      = var.block_public_access
  restrict_public_buckets = var.block_public_access
}
