terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {}

locals {
  service_1_backends = concat(
    [var.shared_backend],
    [
      {
        name    = "unique-origin-1"
        address = "unique1.origin.example.foo.com"
        port    = 80
        comment = "Unique backend for service 1"
      }
    ]
  )

  service_2_backends = [var.shared_backend]
}

resource "fastly_service_cdn_auto" "service_1" {
  name    = var.service_1_name
  comment = "Managed by Terraform"
  domain {
    name = "www.service1.example.com"
  }

  dynamic "backend" {
    for_each = local.service_1_backends
    content {
      name    = backend.value.name
      address = backend.value.address
      port    = backend.value.port
      comment = backend.value.comment
    }
  }

  acl {
    name = "ip_allowlist"
  }

  gzip {
    name          = "default_gzip"
    content_types = ["text/html", "text/css", "application/javascript"]
    extensions    = ["css", "js", "html"]
  }
}

# Image Optimizer must be enabled on service_1 before an
# image_optimizer_default_settings block can be added to fastly_service_cdn_auto.service_1.
# This resource can be applied together with the service's own creation, since it only
# depends on the service's id. But image_optimizer_default_settings is reconciled inside
# the service resource's own create step, which runs before this resource - so the settings
# block itself must be added in a later apply, once this resource has enabled the product.
# See README.md's "Configure Image Optimizer default settings" section.
resource "fastly_service_product_image_optimizer" "service_1" {
  service_id = fastly_service_cdn_auto.service_1.id
}

resource "fastly_service_cdn_auto" "service_2" {
  name    = var.service_2_name
  comment = "Managed by Terraform"
  domain {
    name = "www.service2.example.com"
  }

  dynamic "backend" {
    for_each = local.service_2_backends
    content {
      name    = backend.value.name
      address = backend.value.address
      port    = backend.value.port
      comment = backend.value.comment
    }
  }

  acl {
    name          = "temporary_blocklist"
    force_destroy = true
  }
}
