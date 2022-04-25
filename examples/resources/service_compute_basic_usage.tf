resource "fastly_service_compute" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  package {
    filename = "package.tar.gz"
    source_code_hash = filesha512("package.tar.gz")
  }

  force_destroy = true
}