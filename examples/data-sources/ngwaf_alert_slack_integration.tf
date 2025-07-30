data "fastly_ngwaf_alert_slack_integration" "ngwaf_slack_alerts" {
    workspace_id = fastly_ngwaf_workspace.example.id
}

output "ngwaf_slack_alerts_all" {
  value = data.fastly_ngwaf_alert_slack_integration.ngwaf_slack_alerts
}
