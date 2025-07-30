resource "fastly_ngwaf_alert_jira_integration" "demo_jira_alert" {
  description    = "Some Description"
  key            = "123456789"
  site           = "us1"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}