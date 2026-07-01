resource "fastly_service_cdn_acl_entries" "test" {
  service_id     = fastly_service_cdn.test.id
  acl_id         = fastly_service_acl.test.acl_id
  manage_entries = true

  entry {
    ip = "127.0.0.1"
  }
}
