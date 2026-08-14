logging_splunk {
  name = "{{.LOGGING_SPLUNK_NAME}}"
  url  = "https://splunk.example.com/services/collector/event"
  authentication = {
    token = "test-splunk-token"
  }
  placement = "none"
}
