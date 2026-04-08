terraform {
  required_providers {
    fastly = {
      source  = "example/fastly"
      version = "0.1.0"
    }
  }
}

resource "fastly_service_domain" "this" {
  service_id = var.service_id
  version    = var.service_version
  name       = var.domain_name
}

resource "fastly_service_backend" "this" {
  for_each = { for b in var.backends : b.name => b }

  service_id = var.service_id
  version    = var.service_version
  name       = each.value.name
  address    = each.value.address
  port       = each.value.port
}
