data "fastly_ngwaf_redactions" "ngwaf_redactions" {
    workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}

output "fastly_ngwaf_redactions_all" {
  value = data.fastly_ngwaf_redactions.redactions
}