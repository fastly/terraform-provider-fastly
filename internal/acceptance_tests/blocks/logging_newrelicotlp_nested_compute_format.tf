logging_newrelicotlp {
  name   = "{{.LOGGING_NEWRELIC_NAME}}"
  token  = "test-insert-key"
  format = "%h %l %u %t \"%r\" %>s %b"
}
