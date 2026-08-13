logging_blobstorage {
  name      = "{{.LOGGING_BLOBSTORAGE_NAME}}"
  container = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount"
    sas_token    = "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"
  }
  path              = "/logs/"
  period            = 7200
  gzip_level        = 5
  format            = "%h %l %u %t \"%r\" %>s %b"
  format_version    = 1
  message_type      = "loggly"
  timestamp_format  = "%Y-%m-%dT%H:%M:%S%z"
  processing_region = "us"
  file_max_bytes    = 1048576
}
