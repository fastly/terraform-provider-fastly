resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "cert" {
  key_algorithm   = tls_private_key.key.algorithm
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example.com"
  }

  is_ca_certificate     = true
  validity_period_hours = 360

  allowed_uses = [
    "cert_signing",
    "server_auth",
  ]

  dns_names = ["example.com"]
}

resource "fastly_tls_private_key" "key" {
  key_pem = tls_private_key.key.private_key_pem
  name    = "tf-demo"
}

resource "fastly_tls_certificate" "example" {
  name = "tf-demo"
  certificate_body = tls_self_signed_cert.cert.cert_pem
  depends_on = [fastly_tls_private_key.key] // The private key has to be present before the certificate can be uploaded
}