logging_splunk {
  name = "{{.LOGGING_SPLUNK_NAME_1}}"
  url  = "https://splunk.example.com/services/collector/event"
  authentication = {
    token = "test-splunk-token"
  }
}

logging_splunk {
  name = "{{.LOGGING_SPLUNK_NAME_2}}"
  url  = "https://splunk.example.com/services/collector/event"
  authentication = {
    token = "test-splunk-token"
  }
}
