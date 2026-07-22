resource "fastly_service_logging_newrelicotlp" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_NEWRELIC_NAME}}"
  token      = "test-insert-key"
}
