resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

# Discovered operations depend on traffic and may legitimately be empty.
data "fastly_api_security_discovered_operations" "discovered" {
  service_id = fastly_service_vcl.svc1.id

  # Optional filters
  status = "SAVED"
  method = ["GET"]
  domain = ["api.example.com"]
  path   = "/v1/things"
}

output "api_security_discovered_operations" {
  value = data.fastly_api_security_discovered_operations.discovered.operations
}

output "api_security_discovered_operations_total" {
  value = data.fastly_api_security_discovered_operations.discovered.total
}
