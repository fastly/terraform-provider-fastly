logging_splunk {
  name = "{{.LOGGING_SPLUNK_NAME}}"
  url  = "https://splunk-updated.example.com/services/collector/event"
  authentication = {
    token = "updated-splunk-token"
  }
  use_tls           = true
  processing_region = "eu"
  format_version    = 2
}
