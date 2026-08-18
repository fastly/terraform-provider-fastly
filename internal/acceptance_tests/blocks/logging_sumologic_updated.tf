resource "fastly_service_logging_sumologic" "test" {
  service_id        = fastly_service_cdn.test.id
  version           = {{.SERVICE_VERSION}}
  name              = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url               = "https://collectors.sumologic.com/receiver/v1/http/updated"
  message_type      = "loggly"
  processing_region = "eu"
  format            = "%h %l %u %t \"%r\" %>s %b"
  format_version    = 2
  placement         = "none"
}
