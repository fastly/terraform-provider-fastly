data "fastly_ngwaf_alert_pagerduty_integration" "ngwaf_pagerduty_alerts" {
    workspace_id = fastly_ngwaf_workspace.example.id
}

output "ngwaf_pagerduty_alerts_all" {
  value = data.fastly_ngwaf_alert_pagerduty_integration.ngwaf_pagerduty_alerts
}