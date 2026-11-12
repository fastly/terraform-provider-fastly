resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true

  vcl {
    name    = "my_custom_main_vcl"
    content = file("${path.module}/my_custom_main.vcl")
    main    = true
  }

  vcl {
    name    = "my_custom_library_vcl"
    content = file("${path.module}/my_custom_library.vcl")
  }
}

resource "fastly_domain" "demo" {
  fqdn        = "demo.notexample.com"
  description = "demo"
  service_id  = fastly_service_vcl.demo.id
}
