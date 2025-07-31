resource "fastly_ngwaf_alert_pagerduty_integration" "demo_pagerduty_alert" {
  description    = "Some Description"
  key            = "1234567890abcdef1234567890abcdef"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}