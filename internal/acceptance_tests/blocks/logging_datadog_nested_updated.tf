logging_datadog {
  name = "{{.LOGGING_DATADOG_NAME}}"
  authentication = {
    token = "updated-datadog-key"
  }
  region            = "EU"
  processing_region = "eu"
  format_version    = 2
}
