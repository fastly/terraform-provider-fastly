resource "fastly_acl" "my_acl" {
  name = "My ACL"
  force_destroy = true
}

resource "fastly_acl_entries" "entries" {
  acl_id = fastly_acl.my_acl.acl_id
  force_destroy = true

  entry {
    ip = "127.0.0.1"
    subnet = "24"
    negated = false
    comment = "ACL Entry 1"
  }

  entry {
    ip = "192.168.0.1"
    subnet = "32" 
    negated = true
    comment = "ACL Entry 2"
  }
}