data "fastly_tls_private_key_ids" "demo" {}

data "fastly_tls_private_key" "example" {
  id = fastly_tls_private_key_ids.demo.ids[0]
}
