resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.ACL_NAME}}"
}
