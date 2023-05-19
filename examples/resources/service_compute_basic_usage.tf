variable "hash" {
  type = string
}

resource "fastly_service_compute" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  package {
    filename         = "package.tar.gz"
    source_code_hash = var.hash
  }

  force_destroy = true
}

