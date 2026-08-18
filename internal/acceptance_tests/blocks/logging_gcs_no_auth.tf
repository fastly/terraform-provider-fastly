resource "fastly_service_logging_gcs" "test" {
  service_id  = fastly_service_cdn.test.id
  version     = {{.SERVICE_VERSION}}
  name        = "{{.LOGGING_GCS_NAME}}"
  bucket_name = "fastly-test-bucket"
}
