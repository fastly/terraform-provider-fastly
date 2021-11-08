data "fastly_tls_private_key" "demo" {
  name = "demo-private-key"
}

output "private_key_needs_replacing" {
  value = data.fastly_tls_private_key.demo.replace
}