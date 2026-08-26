resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_api_security_operation_tag" "tag" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "production"
  description = "Production endpoints"
}

resource "fastly_api_security_operation" "example" {
  service_id  = fastly_service_vcl.svc1.id
  method      = "GET"
  domain      = "api.example.com"
  path        = "/v1/things"
  description = "Retrieve things"
  tag_ids     = [fastly_api_security_operation_tag.tag.tag_id]
}
