data "fastly_ngwaf_thresholds" "list_thresholds" {
    workspace_id = fastly_ngwaf_workspace.test_thresholds_workspace.id
}

output "fastly_ngwaf_thesholds_all" {
  value = data.fastly_ngwaf_thresholds.list_thresholds
}