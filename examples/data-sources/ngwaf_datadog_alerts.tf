data "fastly_ngwaf_datadog_alerts" "ngwaf_datadog_alerts" {
    workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}

output "ngwaf_datadog_alerts_all" {
  value = data.ngwaf_datadog_alerts.datadog_alerts
}