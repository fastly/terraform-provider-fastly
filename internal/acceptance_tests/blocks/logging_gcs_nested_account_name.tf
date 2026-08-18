logging_gcs {
  name        = "{{.LOGGING_GCS_NAME}}"
  bucket_name = "fastly-test-bucket"
  project_id  = "fastly-test-project"
  authentication = {
    account_name = "test-service-account"
  }
}
