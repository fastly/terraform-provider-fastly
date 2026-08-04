logging_datadog {
  name = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "test-datadog-key"
  }
  placement = "none"
}
