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

resource "fastly_service_cdn" "example" {
  name = "example-acl-entries"

  domain {
    name = "example.com"
  }

  backend {
    address = "http-me.fastly.dev"
    name    = "backend"
  }
}

resource "fastly_service_cdn_acl" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "example_acl"
}

# Manage ACL entries with explicit resource
resource "fastly_service_cdn_acl_entries" "example" {
  service_id     = fastly_service_cdn.example.id
  acl_id         = fastly_service_cdn_acl.example.acl_id
  manage_entries = true

  entry {
    ip      = "192.0.2.1"
    subnet  = "32"
    negated = false
    comment = "Single IP address"
  }

  entry {
    ip      = "198.51.100.0"
    subnet  = "24"
    negated = false
    comment = "IP range"
  }

  entry {
    ip      = "203.0.113.10"
    subnet  = "32"
    negated = true
    comment = "Negated entry - blocks this IP"
  }
}
