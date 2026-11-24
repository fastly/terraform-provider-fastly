data "fastly_bot_management_workspace" "by_service" {
  service_id = fastly_service_vcl.example.id
}

# Or look it up by its workspace ID.
data "fastly_bot_management_workspace" "by_id" {
  workspace_id = var.workspace_id
}

output "protection_mode" {
  value = data.fastly_bot_management_workspace.by_service.protection_mode
}
