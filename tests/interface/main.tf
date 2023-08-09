terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = ">1.0.0"
    }
  }
}

resource "fastly_service_vcl" "interface-test-project" {
  name = "interface-test-project"

  domain {
    name    = "interface-test-project.fastly-terraform.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}
