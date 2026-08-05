resource "fastly_service_logging_datadog" "test" {
  service_id = fastly_service_compute.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "test-datadog-key"
  }
  format = "%h %l %u %t \"%r\" %>s %b"
}
