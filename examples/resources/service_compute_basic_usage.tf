data "fastly_package_hash" "example" {
  filename = "./path/to/package.tar.gz"
}

resource "fastly_service_compute" "example" {
  name = "demofastly"

  package {
    filename         = "package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}

resource "fastly_domain" "example" {
  fqdn        = "demo.notexample.com"
  description = "demo"
  service_id  = fastly_service_compute.example.id
}
