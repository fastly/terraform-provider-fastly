data "fastly_bot_management_workspace" "example" {
  service_id = fastly_service_vcl.example.id
}

# Block all AI crawlers by default.
resource "fastly_bot_management_category" "ai_crawler" {
  workspace_id = data.fastly_bot_management_workspace.example.id
  category_id  = "cat_ai_crawler"
  action       = "block"
}

# Allow GPTBot, overriding the category's action.
resource "fastly_bot_management_bot" "gptbot" {
  workspace_id = data.fastly_bot_management_workspace.example.id
  category_id  = fastly_bot_management_category.ai_crawler.category_id
  bot_id       = "bot_gptbot"
  action       = "allow"
}
