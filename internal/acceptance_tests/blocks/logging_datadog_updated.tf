resource "fastly_service_logging_datadog" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "updated-datadog-key"
  }
  region            = "EU"
  processing_region = "eu"
  format            = "%h %l %u %t \"%r\" %>s %b"
  format_version    = 2
  placement         = "none"
}
