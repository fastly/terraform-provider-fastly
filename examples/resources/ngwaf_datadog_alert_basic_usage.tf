resource "fastly_ngwaf_datadog_alert" "demo_datadog_alert" {
  description                  = "Some Description"
  integration_key              = "123456789"
  integration_site             = "us1"
  workspace_id                 = fastly_ngwaf_workspace.demo.id
}