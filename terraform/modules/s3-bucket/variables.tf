# Variables for the S3 bucket module

variable "bucket_name" {
  description = "The name of the S3 bucket"
  type        = string
}

variable "prevent_destroy" {
  description = "Prevent accidental deletion of the bucket"
  type        = bool
  default     = false
}

variable "block_public_access" {
  description = "Block all public access to the bucket"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to the bucket"
  type        = map(string)
  default     = {}
}
