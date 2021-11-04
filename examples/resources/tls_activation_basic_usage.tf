resource "fastly_service_v1" "demo" {
  name = "my-service"

  domain {
    name = "example.com"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_private_key" "demo" {
  key_pem = "..."
  name    = "demo-key"
}

resource "fastly_tls_certificate" "demo" {
  certificate_body = "..."
  name             = "demo-cert"
  depends_on       = [fastly_tls_private_key.demo]
}

resource "fastly_tls_activation" "test" {
  certificate_id = fastly_tls_certificate.demo.id
  domain         = "example.com"
  depends_on     = [fastly_service_v1.demo]
}
