resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

# Optional: create an operation (so the data source returns something predictable)
resource "fastly_api_security_operation" "example" {
  service_id  = fastly_service_vcl.svc1.id
  method      = "GET"
  domain      = "api.example.com"
  path        = "/v1/things"
  description = "Retrieve things"
}

# Pagination:
# - The API uses page+limit pagination and returns meta.total and meta.limit.
# - This data source automatically fetches all pages until meta.total is reached.
# - The `limit` argument controls the page size (results per request).
data "fastly_api_security_operations" "ops" {
  service_id = fastly_service_vcl.svc1.id

  # Optional filters
  method = ["GET"]
  domain = ["api.example.com"]
  path   = "/v1/things"

  # Optional page size
  limit = 100

  depends_on = [fastly_api_security_operation.example]
}

output "api_security_operations" {
  value = data.fastly_api_security_operations.ops.operations
}

output "api_security_operations_total" {
  value = data.fastly_api_security_operations.ops.total
}
