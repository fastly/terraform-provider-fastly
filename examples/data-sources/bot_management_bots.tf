data "fastly_bot_management_workspace" "example" {
  service_id = fastly_service_vcl.example.id
}

data "fastly_bot_management_bots" "ai_crawler" {
  workspace_id = data.fastly_bot_management_workspace.example.id
  category_id  = "cat_ai_crawler" # optional; omit to list every bot in the workspace
}

output "ai_crawler_bot_ids" {
  value = [for b in data.fastly_bot_management_bots.ai_crawler.bots : b.bot_id]
}
