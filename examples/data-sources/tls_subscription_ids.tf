data "fastly_tls_subscription_ids" "example" {}

data "fastly_tls_subscription" "example" {
  for_each = data.fastly_tls_subscription_ids.example.ids
  id       = each.value
}

output "subscription_domains" {
  value = [for a in data.fastly_tls_subscription.example : a.certificate_authority]
}