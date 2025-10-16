data "fastly_ngwaf_workspace_rules" "workspace_rules" {
  workspace_id = fastly_ngwaf_workspace.example.id
}

output "fastly_ngwaf_workspace_rules_all" {
  value = data.fastly_ngwaf_workspace_rules.workspace_rules
}
