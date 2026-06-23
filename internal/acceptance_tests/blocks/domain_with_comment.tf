resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.DOMAIN_NAME}}"
  comment    = "{{.DOMAIN_COMMENT}}"
}
