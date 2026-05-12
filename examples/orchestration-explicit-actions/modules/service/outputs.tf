output "service_id" {
  value = fastly_service_vcl_explicit.this.id
}

output "domain" {
  value = fastly_service_domain_explicit.this.name
}

output "backends" {
  value = [for b in fastly_service_backend_explicit.this : b.name]
}
