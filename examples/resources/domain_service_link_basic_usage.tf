resource "fastly_domain_service_link" "example" {
    domain_id = fastly_domain.example.id
    service_id = fastly_service_vcl.example.id
}
