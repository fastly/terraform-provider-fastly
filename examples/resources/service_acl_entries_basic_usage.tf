variable "myacl_name" {
  type = string
  default = "My ACL"
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name = "AWS S3 hosting"
    port = 80
  }

  acl {
    name = var.myacl_name
  }

  force_destroy = true
}

resource "fastly_service_acl_entries" "entries" {
  for_each = {
  for d in fastly_service_vcl.myservice.acl : d.name => d if d.name == var.myacl_name
  }
  service_id = fastly_service_vcl.myservice.id
  acl_id = each.value.acl_id
  entry {
    ip = "127.0.0.1"
    subnet = "24"
    negated = false
    comment = "ACL Entry 1"
  }
}