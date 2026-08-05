resource "fastly_service_logging_datadog" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "test-datadog-key"
  }
}
