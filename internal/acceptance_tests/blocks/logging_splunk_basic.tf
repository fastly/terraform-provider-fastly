resource "fastly_service_logging_splunk" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_SPLUNK_NAME}}"
  url        = "https://splunk.example.com/services/collector/event"
  authentication = {
    token = "test-splunk-token"
  }
}
