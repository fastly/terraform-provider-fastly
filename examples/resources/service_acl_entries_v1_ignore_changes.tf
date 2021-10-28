#...

resource "fastly_service_acl_entries_v1" "entries" {
  for_each   = {
  for d in fastly_service_v1.myservice.acl : d.name => d if d.name == var.myacl_name
  }
  service_id = fastly_service_v1.myservice.id
  acl_id     = each.value.acl_id
  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "ALC Entry 1"
  }

  lifecycle {
    ignore_changes = [entry,]
  }

}