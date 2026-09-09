data "fastly_ai_runtime_control_usage_metrics" "last_week" {
  from = "2026-05-01T00:00:00Z"
  to   = "2026-05-08T00:00:00Z"
}

output "fastly_ai_runtime_control_usage_metrics_all" {
  value = data.fastly_ai_runtime_control_usage_metrics.last_week.usage_metrics
}

output "fastly_ai_runtime_control_total_input_tokens" {
  value = sum([
    for metric in data.fastly_ai_runtime_control_usage_metrics.last_week.usage_metrics :
    metric.quantity if metric.usage_type == "input_tokens"
  ])
}
