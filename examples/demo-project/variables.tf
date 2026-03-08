variable "project_name" {
  type        = string
  description = "Tenant project identifier"
}

variable "environment" {
  type        = string
  default     = "production"
  description = "Environment type"
}

variable "tier" {
  type        = string
  default     = "standard"
  description = "Tenant tier"
}

variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "VPC CIDR block"
}

variable "subdomain" {
  type        = string
  description = "Tenant subdomain"
}

variable "db_instance_class" {
  type        = string
  default     = "db.t3.micro"
  description = "Database instance class"
}
