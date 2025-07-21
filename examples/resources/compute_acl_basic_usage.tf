# IMPORTANT: Deleting a Compute ACL requires first deleting its resource_link.
# This requires a two-step `terraform apply` because we can't guarantee deletion order.

resource "fastly_compute_acl" "example" {
  name = "my_compute_acl"
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
    name        = "my_acl_link"
    resource_id = fastly_compute_acl.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
