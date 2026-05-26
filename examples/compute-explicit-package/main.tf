terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {
  api_token = var.fastly_api_token
}

locals {
  package_path = "${path.module}/${var.package_filename}"
}

resource "fastly_service_compute" "app" {
  name    = var.service_name
  comment = "Managed by Terraform"
}

resource "fastly_service_domain" "app" {
  service_id = fastly_service_compute.app.id
  version    = var.service_version
  name       = var.domain_name
}

resource "fastly_service_backend" "origin" {
  service_id = fastly_service_compute.app.id
  version    = var.service_version
  name       = var.backend.name
  address    = var.backend.address
  port       = var.backend.port
}

# Terraform Actions are not stateful resources, so this terraform_data resource
# provides the stateful diff that triggers package uploads during normal apply.
resource "terraform_data" "compute_package" {
  input = {
    service_id       = fastly_service_compute.app.id
    version          = var.service_version
    filename         = local.package_path
    source_code_hash = filebase64sha256(local.package_path)
  }

  lifecycle {
    action_trigger {
      action = action.fastly_service_compute_package_upload.this
      events = ["create", "update"]
    }
  }
}

action "fastly_service_compute_package_upload" "this" {
  config {
    service_id = fastly_service_compute.app.id
    version    = var.service_version
    filename   = local.package_path
  }
}

# Activation remains explicit. Invoke this only after reviewing the applied
# version and confirming the package upload and versioned resources are ready.
action "fastly_service_version_activate" "production" {
  config {
    service_id = fastly_service_compute.app.id
    version    = var.service_version
    staging    = false
  }
}
