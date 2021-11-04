resource "tls_private_key" "demo" {
  algorithm = "RSA"
}

resource "fastly_tls_private_key" "demo" {
  key_pem = tls_private_key.demo.private_key_pem
  name    = "tf-demo"
}