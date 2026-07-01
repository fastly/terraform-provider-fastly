resource "fastly_service_cdn_acl_entries" "test" {
  service_id     = fastly_service_cdn.test.id
  acl_id         = fastly_service_acl.test.acl_id
  manage_entries = false

  entry {
    ip      = "10.0.0.1"
    subnet  = 24
    negated = false
    comment = "Changed entry"
  }
}
