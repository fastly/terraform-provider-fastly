resource "fastly_acl" "my_acl" {
  name = "My ACL"
  force_destroy = true
}

resource "fastly_acl_entries" "entries" {
  acl_id = fastly_acl.my_acl.acl_id
  force_destroy = true

  entry {
    prefix = "127.0.0.1/32"
    action = "BLOCK"
  }

  entry {
    ip = "192.168.0.1/32"
    action = "ALLOW"
  }
}