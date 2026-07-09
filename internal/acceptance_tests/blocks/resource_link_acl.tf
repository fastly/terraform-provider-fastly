resource "fastly_acl" "acl" {
  name = "{{.ACL_NAME}}"
}

resource "fastly_service_resource_link" "test" {
  service_id  = fastly_service_compute.test.id
  version     = {{.SERVICE_VERSION}}
  name        = "{{.RESOURCE_LINK_NAME}}"
  resource_id = fastly_acl.acl.id
}
