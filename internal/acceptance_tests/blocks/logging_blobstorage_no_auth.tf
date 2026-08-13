resource "fastly_service_logging_blobstorage" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_BLOBSTORAGE_NAME}}"
  container  = "{{.CONTAINER_NAME}}"
}
