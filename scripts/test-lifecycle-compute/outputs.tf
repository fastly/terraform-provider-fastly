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

output "acl_id" {
  value       = fastly_acl.acl.id
  description = "ACL ID"
}

output "acl_name" {
  value       = fastly_acl.acl.name
  description = "ACL name"
}

output "resource_link_id" {
  value       = fastly_service_resource_link.service_1_acl.link_id
  description = "Link ID of the resource_link attaching the ACL to service 1"
}

output "resource_link_version" {
  value       = fastly_service_resource_link.service_1_acl.version
  description = "Service version the resource_link is currently pinned to"
}
