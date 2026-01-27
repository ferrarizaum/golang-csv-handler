# Variables for the S3 folder module

variable "bucket_id" {
  description = "The ID of the S3 bucket where the folder will be created"
  type        = string
}

variable "folder_name" {
  description = "The name of the folder to create (without trailing slash)"
  type        = string
}
