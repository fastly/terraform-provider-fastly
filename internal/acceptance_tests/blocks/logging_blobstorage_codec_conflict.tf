resource "fastly_service_logging_blobstorage" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_BLOBSTORAGE_NAME}}"
  container  = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount"
    sas_token    = "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"
  }
  compression_codec = "gzip"
  gzip_level         = 3
}
