terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {
  # Configure the provider with your Fastly API key
  # api_key = var.fastly_api_key or using FASTLY_API_KEY env var
}

data "fastly_package_hash" "sigil" {
  filename = "/Users/jonmonette/Workspace/Datadotworld/development/fastly-edge-compute/pkg/sigil.tar.gz"
}

# resource "fastly_compute_acls" "blocked_ips" {
#   name          = "blocked_ips"
#   force_destroy = true
# }

# resource "fastly_compute_acls_entries" "entries" {
#   acl_id = fastly_compute_acls.blocked_ips.acl_id
#   force_destroy = true

#   entry {
#     prefix = "127.0.0.1/32"
#     action = "BLOCK"
#   }
# }

data "fastly_compute_acls" "acls" {
    # acls {
    #     name = "blocked_ips"
    # }
}

output "acls" {
  value = data.fastly_compute_acls.acls
}

# Create Compute service with ACLs
resource "fastly_service_compute" "example" {
    # depends_on = [ fastly_compute_acls.blocked_ips ]
  name = "acl_test"

  domain {
    name    = "acl.data.world"
  }

  package {
    filename         = "/Users/jonmonette/Workspace/Datadotworld/development/fastly-edge-compute/pkg/sigil.tar.gz"
    source_code_hash = data.fastly_package_hash.sigil.hash
  }

  # Link ACLs to the compute service

#   resource_link {
#     name        = "blocked_ips"
#     resource_id = fastly_compute_acls.blocked_ips.id
#   }

  force_destroy = true
}





