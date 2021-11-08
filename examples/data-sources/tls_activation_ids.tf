data "fastly_tls_activation_ids" "example" {
  certificate_id = fastly_tls_certificate.example.id
}

data "fastly_tls_activation" "example" {
  for_each = data.fastly_tls_activation_ids.example.ids
  id       = each.value
}

output "activation_domains" {
  value = [for a in data.fastly_tls_activation.example : a.domain]
}