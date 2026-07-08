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

resource "fastly_service_compute" "app" {
  name    = "example-compute-service"
  comment = "Managed by Terraform"
}

# Makes the linked resource available to Wasm code, e.g. as a KV Store or
# Config Store lookup keyed by "store".
resource "fastly_service_resource_link" "store" {
  service_id  = fastly_service_compute.app.id
  version     = 1
  name        = "store"
  resource_id = var.linked_resource_id
}
