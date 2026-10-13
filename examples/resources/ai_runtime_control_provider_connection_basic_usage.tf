resource "fastly_ai_runtime_control_provider_connection" "demo" {
  name     = "OpenAI"
  base_url = "https://api.openai.com"
  api_key  = var.openai_api_key
  models = [
    "gpt-4o",
    "gpt-4o-mini",
  ]
}
