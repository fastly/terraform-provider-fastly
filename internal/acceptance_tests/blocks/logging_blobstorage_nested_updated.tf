logging_blobstorage {
  name      = "{{.LOGGING_BLOBSTORAGE_NAME}}"
  container = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount"
    sas_token    = "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"
  }
  path              = "/updated-logs/"
  period            = 1800
  gzip_level        = 9
  format_version    = 2
  message_type      = "loggly"
  timestamp_format  = "%Y-%m-%dT%H:%M:%S%z"
  processing_region = "eu"
  file_max_bytes    = 2097152
}
