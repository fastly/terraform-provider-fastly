resource "fastly_service_vcl" "example" {
  name = "example-service"

  domain {
    name = "example.com"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains = [for domain in fastly_service_vcl.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}
