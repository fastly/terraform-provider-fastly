resource "fastly_service_compute" "example_with_default_package" {
  name = "demofastly-default-package"

  domain {
    name    = "demo.notexample.com"
    comment = "demo with default package"
  }

  # Package block with neither content nor filename specified
  # This will use the default filename "main.wasm"
  package {
    # source_code_hash can still be specified if needed for updates
  }

  # Must set activate = false when using default package as the file may not exist
  activate      = false
  force_destroy = true
}