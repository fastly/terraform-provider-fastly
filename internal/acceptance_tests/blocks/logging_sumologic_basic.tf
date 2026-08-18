resource "fastly_service_logging_sumologic" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url        = "https://collectors.sumologic.com/receiver/v1/http/test"
}
