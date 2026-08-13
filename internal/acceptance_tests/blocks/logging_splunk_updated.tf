resource "fastly_service_logging_splunk" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_SPLUNK_NAME}}"
  url        = "https://splunk-updated.example.com/services/collector/event"
  authentication = {
    token = "updated-splunk-token"
  }
  tls = {
    ca_cert     = "test-ca-cert"
    client_cert = "test-client-cert"
    client_key  = "test-client-key"
    hostname    = "splunk.example.com"
  }
  use_tls             = true
  processing_region   = "eu"
  request_max_bytes   = 1000000
  request_max_entries = 1000
  format              = "%h %l %u %t \"%r\" %>s %b"
  format_version      = 2
  placement           = "none"
}
