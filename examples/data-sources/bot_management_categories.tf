data "fastly_bot_management_workspace" "example" {
  service_id = fastly_service_vcl.example.id
}

data "fastly_bot_management_categories" "all" {
  workspace_id = data.fastly_bot_management_workspace.example.id
}

output "category_ids" {
  value = [for c in data.fastly_bot_management_categories.all.categories : c.category_id]
}
