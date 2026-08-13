logging_bigquery {
  name       = "{{.LOGGING_BIGQUERY_NAME}}"
  project_id = "fastly-test-project"
  dataset    = "fastly_test_dataset"
  table      = "fastly_test_table"
  authentication = {
    account_name = "test-service-account"
  }
}
