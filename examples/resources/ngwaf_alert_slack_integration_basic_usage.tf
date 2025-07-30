resource "fastly_ngwaf_alert_slack_integration" "demo_slack_alert" {
  description    = "Some Description"
  webhook        = "https://example.com/webhooks/my-service"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}