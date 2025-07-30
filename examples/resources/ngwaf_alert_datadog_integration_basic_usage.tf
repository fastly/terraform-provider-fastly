resource "fastly_ngwaf_alert_datadog_integration" "demo_datadog_alert" {
  description    = "Some Description"
  key            = "123456789"
  site           = "us1"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}