logging_sumologic {
  name              = "{{.LOGGING_SUMOLOGIC_NAME}}"
  url               = "https://collectors.sumologic.com/receiver/v1/http/updated"
  message_type      = "loggly"
  processing_region = "eu"
  format_version    = 2
}
