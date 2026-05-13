output "service_1_id" {
  description = "Fastly service ID for service 1."
  value       = fastly_service_vcl.service_1.id
}

output "service_1_active_version" {
  description = "Currently active version for service 1."
  value       = fastly_service_vcl.service_1.active_version
}

output "service_1_cloned_version" {
  description = "Most recent cloned version tracked by the resource for service 1."
  value       = fastly_service_vcl.service_1.cloned_version
}

output "service_2_id" {
  description = "Fastly service ID for service 2."
  value       = fastly_service_vcl.service_2.id
}

output "service_2_active_version" {
  description = "Currently active version for service 2."
  value       = fastly_service_vcl.service_2.active_version
}

output "service_2_cloned_version" {
  description = "Most recent cloned version tracked by the resource for service 2."
  value       = fastly_service_vcl.service_2.cloned_version
}
