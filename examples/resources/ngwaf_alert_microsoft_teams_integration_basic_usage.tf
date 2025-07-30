resource "fastly_ngwaf_alert_microsoft_teams_integration" "demo_microsoft_teams_alert" {
  description    = "Some Description"
  webhook        = "https://example.com/microsoft-teams/my-service"
  workspace_id   = fastly_ngwaf_workspace.demo.id
}