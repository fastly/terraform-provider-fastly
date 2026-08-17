logging_newrelic {
  name = "{{.LOGGING_NEWRELIC_NAME}}"
  authentication = {
    token = "test-newrelic-key"
  }
  format = "%h %l %u %t \"%r\" %>s %b"
}
