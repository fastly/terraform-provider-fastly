logging_gcs {
  name        = "{{.LOGGING_GCS_NAME}}"
  bucket_name = "fastly-test-bucket"
  authentication = {
    account_name = "test-service-account"
  }
}
