resource "fastly_service_backend" "origin" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.BACKEND_NAME}}"
  address    = "api.example.com"
  port       = 443
  use_ssl    = true
}
