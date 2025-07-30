resource "fastly_ngwaf_alert_jira_integration" "demo_jira_alert" {
  description      = "A description"
  host             = "https://mycompany.atlassian.net"
  issue_type       = "task"
  key              = "a1b2c3d4e5f6789012345678901234567"
  project          = "test"
  username         = "user"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}