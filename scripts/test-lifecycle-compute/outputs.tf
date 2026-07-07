output "service_1_id" {
  value       = fastly_service_compute.service_1.id
  description = "Service 1 ID"
}

output "service_1_name" {
  value       = fastly_service_compute.service_1.name
  description = "Service 1 name"
}

output "service_1_active_version" {
  value       = data.fastly_service_version.service_1.active_version
  description = "Service 1 active version"
}

output "service_1_latest_version" {
  value       = data.fastly_service_version.service_1.latest_version
  description = "Service 1 latest version"
}

output "service_2_id" {
  value       = fastly_service_compute.service_2.id
  description = "Service 2 ID"
}

output "service_2_name" {
  value       = fastly_service_compute.service_2.name
  description = "Service 2 name"
}

output "service_2_active_version" {
  value       = data.fastly_service_version.service_2.active_version
  description = "Service 2 active version"
}

output "service_2_latest_version" {
  value       = data.fastly_service_version.service_2.latest_version
  description = "Service 2 latest version"
}
