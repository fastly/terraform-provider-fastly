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

resource "fastly_acl" "example" {
  name = "example-acl"
}

# Manage ACL entries with explicit resource
resource "fastly_acl_entries" "example" {
  acl_id         = fastly_acl.example.id
  manage_entries = true

  entries = {
    "192.0.2.0/24"    = "ALLOW"
    "198.51.100.0/24" = "BLOCK"
    "203.0.113.10/32" = "BLOCK"
  }
}
