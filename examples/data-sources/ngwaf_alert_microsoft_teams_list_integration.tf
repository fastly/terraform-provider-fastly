data "fastly_ngwaf_alert_microsoft_teams_integration" "ngwaf_microsoft_teams_alerts" {
    workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}

output "ngwaf_microsoft_teams_alerts_all" {
  value = data.ngwaf_microsoft_teams_alerts.microsoft_teams_alerts
}