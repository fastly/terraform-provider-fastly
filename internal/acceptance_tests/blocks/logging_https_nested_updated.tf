logging_https {
  name               = "{{.LOGGING_HTTPS_NAME}}"
  url                = "https://https-updated.example.com/logs"
  gzip_level         = 6
  processing_region  = "eu"
  format_version     = 2
}
