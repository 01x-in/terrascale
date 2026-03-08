variable "project_name" {
  type        = string
  description = "Unique project identifier for this deployment."
}

variable "subdomain" {
  type        = string
  description = "Subdomain prefix. Use an empty string for the apex domain."
  default     = ""
}

variable "environment" {
  type        = string
  description = "Environment name for tagging and naming."
  default     = "production"
}

variable "domain_name" {
  type        = string
  description = "Base domain name for the site."
  default     = "terrascale.link"
}

variable "hosted_zone_name" {
  type        = string
  description = "Route53 hosted zone name."
  default     = "terrascale.link"
}

variable "aws_region" {
  type        = string
  description = "Primary AWS region for S3 and Route53 resources."
  default     = "us-east-2"
}

variable "tags" {
  type        = map(string)
  description = "Additional tags to apply to all supported resources."
  default     = {}
}
