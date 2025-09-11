variable "myacl_name" {
  type    = string
  default = "My ACL"
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  acl {
    name = var.myacl_name
  }
}

resource "fastly_service_acl_entries" "entries" {
  service_id = fastly_service_vcl.myservice.id
  acl_id     = { for d in fastly_service_vcl.myservice.acl : d.name => d.acl_id }[var.myacl_name]
  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "ACL Entry 1"
  }
}
