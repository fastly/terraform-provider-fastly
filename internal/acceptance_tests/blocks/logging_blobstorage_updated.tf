resource "fastly_service_logging_blobstorage" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_BLOBSTORAGE_NAME}}"
  container  = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount2"
    sas_token    = "sv=2021-08-06&sr=b&sig=A%2Fx8h5vQ3ZuTn2R9tYkX7wL0mCq1oPzB9dFsEjKa4Uc%3D&se=2051-01-01T00%3A00%3A00Z&sp=rw"
  }
  path              = "/updated-logs/"
  period            = 1800
  gzip_level        = 9
  format_version    = 2
  message_type      = "loggly"
  timestamp_format  = "%Y-%m-%dT%H:%M:%S%z"
  processing_region = "eu"
  file_max_bytes    = 2097152
  public_key        = trimspace(file("{{.PUBLIC_KEY_PATH}}"))
}
