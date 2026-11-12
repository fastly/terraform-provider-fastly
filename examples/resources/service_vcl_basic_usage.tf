resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}

resource "fastly_domain" "demo" {
  fqdn        = "demo.notexample.com"
  description = "demo"
  service_id  = fastly_service_vcl.demo.id
}
