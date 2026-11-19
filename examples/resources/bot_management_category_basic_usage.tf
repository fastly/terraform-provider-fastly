data "fastly_bot_management_workspace" "example" {
  service_id = fastly_service_vcl.example.id
}

# Block every AI crawler that doesn't have its own action.
resource "fastly_bot_management_category" "ai_crawler" {
  workspace_id = data.fastly_bot_management_workspace.example.id
  category_id  = "cat_ai_crawler"
  action       = "block"
}
