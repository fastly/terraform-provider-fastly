# IMPORTANT: Deleting a Config Store requires first deleting its resource_link.
# This requires a two-step `terraform apply` as we can't guarantee deletion order.
# e.g. resource_link deletion within fastly_service_compute might not finish first.
resource "fastly_configstore" "example" {
  name = "%s"
}

resource "fastly_configstore_entries" "example" {
  store_id = fastly_configstore.example.id
  entries = {
    key1 : "value1"
    key2 : "value2"
  }
  manage_entries = true
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
    resource_id = fastly_configstore.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
