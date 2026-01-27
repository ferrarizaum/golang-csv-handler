# S3 Folder Module
# This module creates a folder (prefix) in an S3 bucket

resource "aws_s3_object" "folder" {
  bucket = var.bucket_id
  key    = "${var.folder_name}/"
}
