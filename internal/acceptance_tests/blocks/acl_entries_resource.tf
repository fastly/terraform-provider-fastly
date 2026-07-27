resource "fastly_acl" "acl" {
  name = "{{.ACL_NAME}}"
}

resource "fastly_acl_entries" "acl_entries" {
  acl_id         = fastly_acl.acl.id
  entries        = {{.ENTRIES}}
  manage_entries = true
}
