resource "fastly_ngwaf_alert_opsgenie_integration" "demo_opsgenie_alert" {
  description      = "A description"
  key              = "123456789"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}