resource "fastly_service_logging_https" "test" {
  service_id = fastly_service_compute.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_HTTPS_NAME}}"
  url        = "https://https.example.com/logs"
}
