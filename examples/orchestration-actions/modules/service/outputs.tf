output "service_id" {
  value = fastly_service.this.id
}

output "domain" {
  value = fastly_service_domain.this.name
}

output "backends" {
  value = [for b in fastly_service_backend.this : b.name]
}
