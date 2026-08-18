logging_sumologic {
  name   = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url    = "https://collectors.sumologic.com/receiver/v1/http/test"
  format = "%h %l %u %t \"%r\" %>s %b"
}
