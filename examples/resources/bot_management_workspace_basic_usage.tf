data "fastly_bot_management_workspace" "example" {
  service_id = fastly_service_vcl.example.id
}

resource "fastly_bot_management_workspace" "example" {
  workspace_id    = data.fastly_bot_management_workspace.example.id
  protection_mode = "block"
}
