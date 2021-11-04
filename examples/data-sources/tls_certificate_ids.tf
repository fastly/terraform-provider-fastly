data "fastly_tls_certificate_ids" "example" {}

resource "fastly_tls_activation" "example" {
  certificate_id = data.fastly_tls_certificate_ids.example.ids[0]
  // ...
}