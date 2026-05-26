terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

resource "fastly_service_cdn" "this" {
  name    = var.service_name
  comment = "Managed by Terraform demo provider"
}

resource "fastly_service_domain" "this" {
  service_id = fastly_service_cdn.this.id
  version    = var.service_version
  name       = var.domain_name
}

resource "fastly_service_backend" "this" {
  for_each = { for b in var.backends : b.name => b }

  service_id = fastly_service_cdn.this.id
  version    = var.service_version
  name       = each.value.name
  address    = each.value.address
  port       = each.value.port
}
