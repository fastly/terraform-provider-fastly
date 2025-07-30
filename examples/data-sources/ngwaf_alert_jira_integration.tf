data "fastly_ngwaf_alert_jira_integration" "ngwaf_jira_alerts" {
    workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}

output "ngwaf_jira_alerts_all" {
  value = data.ngwaf_jira_alerts.jira_alerts
}