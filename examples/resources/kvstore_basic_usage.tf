resource "fastly_kvstore" "example" {
  name = "my_kv_store"

  # Provide a service to link the KV Store to.
  resource_link {
    service_id      = fastly_service_compute.example.id
    service_version = 1
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

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
