data "fastly_tls_platform_certificate_ids" "example" {}

data "fastly_tls_platform_certificate" "example" {
  id = data.fastly_tls_platform_certificate_ids.example.ids[0]
}