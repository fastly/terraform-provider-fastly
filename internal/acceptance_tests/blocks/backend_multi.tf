resource "fastly_service_backend" "origin1" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.BACKEND_1_NAME}}"
  address    = "api1.example.com"
  port       = 443
  use_ssl    = true
}

resource "fastly_service_backend" "origin2" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.BACKEND_2_NAME}}"
  address    = "api2.example.com"
  port       = 443
  use_ssl    = true
}
