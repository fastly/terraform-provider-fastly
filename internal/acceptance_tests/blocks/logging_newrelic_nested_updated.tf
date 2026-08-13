logging_newrelic {
  name = "{{.LOGGING_NEWRELIC_NAME}}"
  authentication = {
    token = "updated-newrelic-key"
  }
  region            = "EU"
  processing_region = "eu"
  format_version    = 2
}
