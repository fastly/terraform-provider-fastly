resource "fastly_service_logging_newrelic" "test" {
  service_id = fastly_service_compute.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_NEWRELIC_NAME}}"
  authentication = {
    token = "test-newrelic-key"
  }
}
