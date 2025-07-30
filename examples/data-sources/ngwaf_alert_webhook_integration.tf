data "fastly_ngwaf_alert_webhook_integration" "ngwaf_webhook_alerts" {
    workspace_id = fastly_ngwaf_workspace.example.id
}

output "ngwaf_webhook_alerts_all" {
  value = data.fastly_ngwaf_alert_webhook_integration.ngwaf_webhook_alerts
}
