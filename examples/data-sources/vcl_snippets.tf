resource "fastly_service_vcl" "example" {
  name = "Example Service"

  domain {
    name = "example.com"
  }

  force_destroy = true
}

data "fastly_vcl_snippets" "example" {
  service_id      = fastly_service_vcl.example.id
  service_version = fastly_service_vcl.example.active_version
}

output "service_vcl_snippets" {
  value = data.fastly_vcl_snippets.example
}
