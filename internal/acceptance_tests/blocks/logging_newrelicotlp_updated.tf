resource "fastly_service_logging_newrelicotlp" "test" {
  service_id        = fastly_service_cdn.test.id
  version           = {{.SERVICE_VERSION}}
  name              = "{{.LOGGING_NEWRELIC_NAME}}"
  token             = "updated-insert-key"
  region            = "EU"
  url               = "https://otlp.eu01.nr-data.net"
  processing_region = "eu"
  format            = "%h %l %u %t \"%r\" %>s %b"
  format_version    = 2
  placement         = "none"
}
