resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-example"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_api_security_operation_tag" "example" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "production"
  description = "Tag for production endpoints"
}
