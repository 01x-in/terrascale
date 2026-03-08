resource "local_file" "tenant_config" {
  content = jsonencode({
    project_name      = var.project_name
    environment       = var.environment
    tier              = var.tier
    vpc_cidr          = var.vpc_cidr
    subdomain         = var.subdomain
    db_instance_class = var.db_instance_class
    provisioned_at    = timestamp()
  })
  filename = "${path.module}/.tenant-data/${var.project_name}/config.json"
}

resource "random_password" "api_key" {
  length  = 32
  special = false
}

resource "random_id" "db_suffix" {
  byte_length = 4
}
