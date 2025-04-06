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
    resource_id = fastly_acl.example.id
  }

  force_destroy = true
}

resource "fastly_acl" "example" {
  name          = "example_acl"
  force_destroy = true
}