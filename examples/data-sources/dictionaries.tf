resource "fastly_service_vcl" "example" {
  name = "Example Service"

  domain {
    name = "example.com"
  }

  dictionary {
    name = "example_1"
  }

  dictionary {
    name = "example_2"
  }

  dictionary {
    name = "example_3"
  }

  force_destroy = true
}

data "fastly_dictionaries" "example" {
  depends_on      = [fastly_service_vcl.example]
  service_id      = fastly_service_vcl.example.id
  service_version = fastly_service_vcl.example.active_version
}

output "service_dictionaries" {
  value = data.fastly_dictionaries.example
}
