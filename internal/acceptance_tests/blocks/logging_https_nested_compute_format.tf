logging_https {
  name   = "{{.LOGGING_HTTPS_NAME}}"
  url    = "https://https.example.com/logs"
  format = "%h %l %u %t \"%r\" %>s %b"
}
