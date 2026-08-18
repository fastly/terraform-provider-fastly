logging_sumologic {
  name      = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url       = "https://collectors.sumologic.com/receiver/v1/http/test"
  placement = "none"
}
