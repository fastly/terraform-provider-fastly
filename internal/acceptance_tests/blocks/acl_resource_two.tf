resource "fastly_acl" "acl1" {
  name = "{{.ACL_NAME_1}}"
}

resource "fastly_acl" "acl2" {
  name = "{{.ACL_NAME_2}}"
}
