logging_datadog {
  name = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "test-datadog-key"
  }
  format = "%h %l %u %t \"%r\" %>s %b"
}
