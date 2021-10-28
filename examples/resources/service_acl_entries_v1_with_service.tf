variable "myacl_name" {
  type    = string
  default = "My ACL"
}

resource "fastly_service_v1" "myservice" {
  #...
  acl {
    name = var.myacl_name
  }
  #...
}

resource "fastly_service_acl_entries_v1" "entries" {
  service_id = fastly_service_v1.myservice.id
  acl_id     = {for d in fastly_service_v1.myservice.acl : d.name => d.acl_id}[
  var.myacl_name
  ]
  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "ALC Entry 1"
  }
}