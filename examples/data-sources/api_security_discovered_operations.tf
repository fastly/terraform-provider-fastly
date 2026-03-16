resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

# Discovered operations depend on traffic and may legitimately be empty.
#
# Pagination:
# - The API uses page+limit pagination and returns meta.total and meta.limit.
# - This data source automatically fetches all pages until meta.total is reached.
# - The `limit` argument controls the page size (results per request).
data "fastly_api_security_discovered_operations" "discovered" {
  service_id = fastly_service_vcl.svc1.id

  # Optional filters
  status = "SAVED"
  method = ["GET"]
  domain = ["api.example.com"]
  path   = "/v1/things"

  # Optional page size
  limit = 100
}

output "api_security_discovered_operations" {
  value = data.fastly_api_security_discovered_operations.discovered.operations
}

output "api_security_discovered_operations_total" {
  value = data.fastly_api_security_discovered_operations.discovered.total
}
