output "service_id" {
  description = "The Fastly Compute service ID."
  value       = fastly_service_compute_explicit.app.id
}

output "service_version" {
  description = "The service version targeted by this example."
  value       = var.service_version
}

output "package_hash" {
  description = "Hash used to trigger package uploads during terraform apply."
  value       = terraform_data.compute_package.input.source_code_hash
}
