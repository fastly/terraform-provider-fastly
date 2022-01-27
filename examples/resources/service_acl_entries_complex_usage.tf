locals {
  acl_name = "my_acl"
  acl_entries = [
    {
      ip      = "1.2.3.4"
      comment = "acl_entry_1"
    },
    {
      ip      = "1.2.3.5"
      comment = "acl_entry_2"
    },
    {
      ip      = "1.2.3.6"
      comment = "acl_entry_3"
    }
  ]
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "1.2.3.4"
    name    = "localhost"
    port    = 80
  }

  acl {
    name = local.acl_name
  }

  force_destroy = true
}

resource "fastly_service_acl_entries" "entries" {
  for_each = {
  for d in fastly_service_vcl.myservice.acl : d.name => d if d.name == local.acl_name
  }
  service_id = fastly_service_vcl.myservice.id
  acl_id = each.value.acl_id
  dynamic "entry" {
    for_each = [for e in local.acl_entries : {
      ip      = e.ip
      comment = e.comment
    }]

    content {
      ip      = entry.value.ip
      subnet  = 22
      comment = entry.value.comment
      negated = false
    }
  }
}