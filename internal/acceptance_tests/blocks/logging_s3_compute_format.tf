resource "fastly_service_logging_s3" "test" {
  service_id  = fastly_service_compute.test.id
  version     = {{.SERVICE_VERSION}}
  name        = "{{.LOGGING_S3_NAME}}"
  bucket_name = "{{.BUCKET_NAME}}"
  authentication = {
    access_key = "AKIAIOSFODNN7EXAMPLE"
    secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  format = "%h %l %u %t \"%r\" %>s %b"
}
