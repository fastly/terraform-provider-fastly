data "fastly_ngwaf_alert_opsgenie_integration" "ngwaf_opsgenie_alerts" {
    workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}

output "ngwaf_opsgenie_alerts_all" {
  value = data.ngwaf_opsgenie_alerts.opsgenie_alerts
}