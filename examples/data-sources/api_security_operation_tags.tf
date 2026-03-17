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

# Pagination:
# - The API uses page+limit pagination and returns meta.total and meta.limit.
# - This data source automatically fetches all pages until meta.total is reached.
# - The `limit` argument controls the page size (results per request).
data "fastly_api_security_operation_tags" "tags" {
  service_id = fastly_service_vcl.svc1.id

  # Optional page size
  limit = 100

  depends_on = [fastly_api_security_operation_tag.example]
}

output "api_security_operation_tags" {
  value = data.fastly_api_security_operation_tags.tags.tags
}

output "api_security_operation_tags_total" {
  value = data.fastly_api_security_operation_tags.tags.total
}
