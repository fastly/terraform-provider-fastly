data "fastly_package_hash" "example" {
  filename = "./path/to/package.tar.gz"
}

resource "fastly_service_compute" "example" {
  # ...

  package {
    filename         = "./path/to/package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
}
