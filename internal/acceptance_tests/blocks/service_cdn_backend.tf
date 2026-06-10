resource "fastly_service_backend" "test" {
  service_id        = fastly_service_cdn.test.id
  version           = {{.SERVICE_VERSION}}
  name              = "{{.BACKEND_NAME}}"
  address           = "api.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "api.example.com"
  ssl_sni_hostname  = "api.example.com"
}
