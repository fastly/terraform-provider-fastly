resource "fastly_service_backend" "origin" {
  service_id            = fastly_service_cdn.test.id
  version               = {{.SERVICE_VERSION}}
  name                  = "{{.BACKEND_NAME}}"
  address               = "api.example.com"
  port                  = 443
  use_ssl               = true
  ssl_check_cert        = false
  ssl_cert_hostname     = "cert.example.com"
  ssl_sni_hostname      = "sni.example.com"
  min_tls_version       = "1.2"
  max_tls_version       = "1.3"
  weight                = 75
  max_conn              = 100
  connect_timeout       = 2000
  first_byte_timeout    = 10000
  between_bytes_timeout = 5000
  auto_loadbalance      = true
  comment               = "Full test backend"
}
