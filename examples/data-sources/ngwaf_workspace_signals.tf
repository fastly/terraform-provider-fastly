data "fastly_ngwaf_workspace_signals" "workspace_signals" {
  workspace_id = fastly_ngwaf_workspace.example.id
}

output "fastly_ngwaf_workspace_signals_all" {
  value = data.fastly_ngwaf_workspace_signals.workspace_signals
}
