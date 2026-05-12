output "service_1_id" {
  description = "The Fastly service ID for service 1"
  value       = module.service_1.service_id
}

output "service_2_id" {
  description = "The Fastly service ID for service 2"
  value       = module.service_2.service_id
}
