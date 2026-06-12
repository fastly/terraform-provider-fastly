action "fastly_service_version_clone" "test" {
  config {
    service_id = fastly_service_cdn.test.id
    version    = {{.SERVICE_VERSION}}
  }
}
