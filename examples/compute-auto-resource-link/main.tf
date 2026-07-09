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

resource "fastly_service_compute_auto" "app" {
  name          = "example-compute-auto-service"
  force_destroy = true

  domain {
    name = "example-compute-auto.edgecompute.app"
  }

  # Makes the linked resource available to Wasm code, e.g. as a KV Store or
  # Config Store lookup keyed by "store".
  resource_link {
    name        = "store"
    resource_id = var.linked_resource_id
  }

  package {
    filename = "${path.module}/pkg/package.tar.gz"
  }
}
