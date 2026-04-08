output "service_1_id" {
  description = "The Fastly service ID for service 1"
  value       = fastly_service.service_1.id
}

output "service_2_id" {
  description = "The Fastly service ID for service 2"
  value       = fastly_service.service_2.id
}
