resource "fastly_ai_runtime_control_virtual_key" "demo" {
  name          = "prod website chatbot"
  model         = "claude-sonnet-4-20250514"
  provider_name = "Anthropic"
  expires_at    = "2026-08-05T17:13:36Z"
}
