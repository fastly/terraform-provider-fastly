---
layout: "fastly"
page_title: "Fastly: secretstore_entries"
sidebar_current: "docs-fastly-resource-secretstore-entries"
description: |-
  A secret within a Secret Store.
---

# fastly_secretstore_entries

The Secret Store (`fastly_secretstore`) can be seeded with initial key-value pairs using the `fastly_secretstore_entries` resource.

After the first `terraform apply` the default behaviour is to ignore any further configuration changes to those key-value pairs. Terraform will expect modifications to happen outside of Terraform (e.g. new key-value pairs to be managed using the [Fastly API](https://developer.fastly.com/reference/api/) or [Fastly CLI](https://developer.fastly.com/learning/tools/cli/)).

To change the default behaviour (so Terraform continues to manage the key-value pairs within the configuration) set `manage_entries = true`.

~> **Note:** Terraform should not be used to store large amounts of data, so it's recommended you leave the default behaviour in place and only seed the store with a small amount of key-value pairs. For more information see ["Configuration not data"](https://developer.fastly.com/learning/integrations/orchestration/terraform/#configuration-not-data).

## Example Usage

Basic usage (with seeded values):

```terraform
# IMPORTANT: Deleting a Secret Store requires first deleting its resource_link.
# This requires a two-step `terraform apply` as we can't guarantee deletion order.
# e.g. resource_link deletion within fastly_service_compute might not finish first.
resource "fastly_secretstore" "example" {
  name = "my_secret_store"
}

# NOTE: When running `terraform apply` make sure to first export env vars.
#
# export TF_VAR_key1=<SECRET>
# export TF_VAR_key2=<SECRET>
#
# These will correspond to the following input variables.

variable "key1" {
  description = "The secret for Key 1"
  type        = string
  sensitive   = true
}

variable "key2" {
  description = "The secret for Key 2"
  type        = string
  sensitive   = true
}

# NOTE: The Fastly Secret Store API expects values to be base64 encoded.
resource "fastly_secretstore_entries" "example" {
  store_id = fastly_secretstore.example.id
  entries = {
    key1 : base64encode(var.key1)
    key2 : base64encode(var.key2)
  }
}

resource "fastly_service_compute" "example" {
  name = "my_compute_service"

  domain {
    name = "demo.example.com"
  }

  package {
    filename         = "package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  resource_link {
    name        = "my_resource_link"
    resource_id = fastly_secretstore.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
```

To have Terraform manage the initially seeded key-value pairs defined in your configuration, then you must set `manage_entries = true` (this will cause any key-value pairs added outside of Terraform to be deleted):

```terraform
# IMPORTANT: Deleting a Secret Store requires first deleting its resource_link.
# This requires a two-step `terraform apply` as we can't guarantee deletion order.
# e.g. resource_link deletion within fastly_service_compute might not finish first.
resource "fastly_secretstore" "example" {
  name = "my_secret_store"
}

# NOTE: When running `terraform apply` make sure to first export env vars.
#
# export TF_VAR_key1=<SECRET>
# export TF_VAR_key2=<SECRET>
#
# These will correspond to the following input variables.

variable "key1" {
  description = "The secret for Key 1"
  type        = string
  sensitive   = true
}

variable "key2" {
  description = "The secret for Key 2"
  type        = string
  sensitive   = true
}

# NOTE: The Fastly Secret Store API expects values to be base64 encoded.
resource "fastly_secretstore_entries" "example" {
  store_id = fastly_secretstore.example.id
  entries = {
    key1 : base64encode(var.key1)
    key2 : base64encode(var.key2)
  }
  manage_entries = true # force Terraform to ignore external changes
}

resource "fastly_service_compute" "example" {
  name = "my_compute_service"

  domain {
    name = "demo.example.com"
  }

  package {
    filename         = "package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  resource_link {
    name        = "my_resource_link"
    resource_id = fastly_secretstore.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
```
