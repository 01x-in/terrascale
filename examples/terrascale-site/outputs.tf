output "site_url" {
  description = "Primary URL for the deployed site."
  value       = "https://${var.subdomain != "" ? "${var.subdomain}.${var.domain_name}" : var.domain_name}"
}

output "cloudfront_url" {
  description = "CloudFront distribution domain."
  value       = "https://${aws_cloudfront_distribution.site.domain_name}"
}

output "s3_bucket" {
  description = "Bucket containing the static site assets."
  value       = aws_s3_bucket.site.bucket
}

output "certificate_arn" {
  description = "ACM certificate used for HTTPS."
  value       = aws_acm_certificate_validation.site.certificate_arn
}
