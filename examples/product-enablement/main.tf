terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {
  # API token set via FASTLY_API_TOKEN environment variable
}

resource "fastly_service_cdn_auto" "example" {
  name = "example-service"

  domain {
    name = "example.com"
  }

  backend {
    name    = "example-backend"
    address = "example.com"
    port    = 443
    shield  = "amsterdam-nl" # required for image_optimizer
  }

  force_destroy = true
}

# Each product is its own resource: creating it enables the product on
# service_id, destroying it disables the product. There's no separate
# "enabled" attribute to toggle.

resource "fastly_product_enablement_brotli_compression" "example" {
  service_id = fastly_service_cdn_auto.example.id
}

resource "fastly_product_enablement_image_optimizer" "example" {
  service_id = fastly_service_cdn_auto.example.id
}

resource "fastly_product_enablement_domain_inspector" "example" {
  service_id = fastly_service_cdn_auto.example.id
}

resource "fastly_product_enablement_websockets" "example" {
  service_id = fastly_service_cdn_auto.example.id
}

resource "fastly_product_enablement_ddos_protection" "example" {
  service_id = fastly_service_cdn_auto.example.id
  mode       = "log"
}

# Applying the same product to multiple services with for_each:

variable "service_ids" {
  type    = set(string)
  default = []
}

resource "fastly_product_enablement_domain_inspector" "fleet" {
  for_each = var.service_ids

  service_id = each.value
}
