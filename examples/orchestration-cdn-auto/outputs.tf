output "service_1_id" {
  description = "Fastly service ID for service 1."
  value       = fastly_service_cdn_auto.service_1.id
}

output "service_1_active_version" {
  description = "Currently active version for service 1."
  value       = fastly_service_cdn_auto.service_1.active_version
}

output "service_1_managed_version" {
  description = "Most recent service version managed by the resource for service 1."
  value       = fastly_service_cdn_auto.service_1.managed_version
}

output "service_2_id" {
  description = "Fastly service ID for service 2."
  value       = fastly_service_cdn_auto.service_2.id
}

output "service_2_active_version" {
  description = "Currently active version for service 2."
  value       = fastly_service_cdn_auto.service_2.active_version
}

output "service_2_managed_version" {
  description = "Most recent service version managed by the resource for service 2."
  value       = fastly_service_cdn_auto.service_2.managed_version
}

output "service_1_acl_id" {
  description = "ACL ID for service 1 IP allowlist."
  value       = fastly_service_cdn_auto.service_1.acl[0].acl_id
}

output "service_2_acl_id" {
  description = "ACL ID for service 2 temporary blocklist."
  value       = fastly_service_cdn_auto.service_2.acl[0].acl_id
}
