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

data "fastly_api_security_operations" "ops" {
  service_id = fastly_service_vcl.svc1.id

  # Optional filters
  method = ["GET"]
  domain = ["api.example.com"]
  path   = "/v1/things"

  depends_on = [fastly_api_security_operation.example]
}

output "api_security_operations" {
  value = data.fastly_api_security_operations.ops.operations
}

output "api_security_operations_total" {
  value = data.fastly_api_security_operations.ops.total
}
