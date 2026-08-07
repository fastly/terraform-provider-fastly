resource "fastly_service_logging_bigquery" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.LOGGING_BIGQUERY_NAME}}"
  project_id = "fastly-test-project"
  dataset    = "fastly_test_dataset"
  table      = "fastly_test_table"
}
