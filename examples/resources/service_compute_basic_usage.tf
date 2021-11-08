resource "fastly_service_compute" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  package {
    filename = "package.tar.gz"
    source_code_hash = filesha512("package.tar.gz")
  }

  force_destroy = true
}