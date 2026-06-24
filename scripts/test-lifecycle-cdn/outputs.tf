output "service_1_id" {
  description = "Service 1 ID"
  value       = fastly_service_cdn.service_1.id
}

output "service_1_active_version" {
  description = "Service 1 active version"
  value       = data.fastly_service_version.service_1.active_version
}

output "service_1_latest_version" {
  description = "Service 1 latest version"
  value       = data.fastly_service_version.service_1.latest_version
}

output "service_2_id" {
  description = "Service 2 ID"
  value       = fastly_service_cdn.service_2.id
}

output "service_1_acl_id" {
  description = "Service 1 ACL ID"
  value       = fastly_service_acl.service_1_acl.acl_id
}

output "service_2_acl_id" {
  description = "Service 2 ACL ID"
  value       = fastly_service_acl.service_2_acl.acl_id
}

output "service_2_active_version" {
  description = "Service 2 active version"
  value       = data.fastly_service_version.service_2.active_version
}

output "service_2_latest_version" {
  description = "Service 2 latest version"
  value       = data.fastly_service_version.service_2.latest_version
}
