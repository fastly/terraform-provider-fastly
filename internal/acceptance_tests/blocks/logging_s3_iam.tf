resource "fastly_service_logging_s3" "test" {
  service_id  = fastly_service_cdn.test.id
  version     = {{.SERVICE_VERSION}}
  name        = "{{.LOGGING_S3_NAME}}"
  bucket_name = "{{.BUCKET_NAME}}"
  authentication = {
    iam_role = "arn:aws:iam::123456789012:role/FastlyS3Access"
  }
}
