data "fastly_ngwaf_workspace_lists" "workspace_lists" {
  workspace_id = fastly_ngwaf_workspace.example.id
}

output "fastly_ngwaf_workspace_lists_all" {
  value = data.fastly_ngwaf_workspace_lists.workspace_lists
}
