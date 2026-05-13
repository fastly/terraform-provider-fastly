terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

resource "fastly_service_domain_explicit" "this" {
  service_id = var.service_id
  version    = var.service_version
  name       = var.domain_name
}

resource "fastly_service_backend_explicit" "this" {
  for_each = { for b in var.backends : b.name => b }

  service_id = var.service_id
  version    = var.service_version
  name       = each.value.name
  address    = each.value.address
  port       = each.value.port
}
