resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

# Optional: create a tag so the data source returns a predictable result
resource "fastly_api_security_operation_tag" "example" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "example-tag"
  description = "Example tag"
}

# Note:
# The /tags endpoint currently returns all tags and does not support pagination
# or a limit parameter (query args like ?limit=... are ignored).
data "fastly_api_security_operation_tags" "tags" {
  service_id = fastly_service_vcl.svc1.id

  depends_on = [fastly_api_security_operation_tag.example]
}

output "api_security_operation_tags" {
  value = data.fastly_api_security_operation_tags.tags.tags
}

output "api_security_operation_tags_total" {
  value = data.fastly_api_security_operation_tags.tags.total
}
