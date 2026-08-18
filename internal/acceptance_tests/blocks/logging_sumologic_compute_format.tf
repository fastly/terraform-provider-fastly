resource "fastly_service_logging_sumologic" "test" {
  service_id = fastly_service_compute.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url        = "https://collectors.sumologic.com/receiver/v1/http/test"
  format     = "%h %l %u %t \"%r\" %>s %b"
}
