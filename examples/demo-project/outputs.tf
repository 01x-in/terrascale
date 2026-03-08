output "vpc_id" {
  value = "vpc-demo-${var.project_name}"
}

output "db_endpoint" {
  value = "${var.project_name}-${random_id.db_suffix.hex}.db.example.com"
}

output "db_name" {
  value = "${replace(var.project_name, "-", "_")}_db"
}

output "s3_bucket" {
  value = "${var.project_name}-uploads"
}

output "api_endpoint" {
  value = "https://api.${var.subdomain}.example.com"
}

output "frontend_url" {
  value = "https://${var.subdomain}.example.com"
}
