action "fastly_service_version_activate" "test" {
  config {
    service_id = fastly_service_cdn.test.id
    version    = {{.SERVICE_VERSION}}
  }
}
