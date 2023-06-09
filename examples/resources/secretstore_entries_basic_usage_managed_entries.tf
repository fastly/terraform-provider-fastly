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
