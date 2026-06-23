resource "fastly_service_domain" "primary" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.DOMAIN_1_NAME}}"
}

resource "fastly_service_domain" "additional" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.DOMAIN_2_NAME}}"
  comment    = "Additional domain"
}
