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
  name     = var.service_1_name
  comment  = "Managed by Terraform"
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
}

resource "fastly_service_cdn_auto" "service_2" {
  name     = var.service_2_name
  comment  = "Managed by Terraform"
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
}
