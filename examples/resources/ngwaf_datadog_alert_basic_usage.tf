resource "fastly_ngwaf_datadog_alert" "demo_datadog_alert" {
  description    = "Some Description"
  key            = "123456789"
  site           = "us1"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}