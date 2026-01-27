# Outputs for the S3 folder module

output "folder_key" {
  description = "The key (path) of the created folder"
  value       = aws_s3_object.folder.key
}

output "folder_id" {
  description = "The ID of the S3 object"
  value       = aws_s3_object.folder.id
}
