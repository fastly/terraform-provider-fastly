data "fastly_ai_runtime_control_virtual_keys" "anthropic" {
  provider_name = "Anthropic"
}

output "fastly_ai_runtime_control_virtual_keys_all" {
  value = data.fastly_ai_runtime_control_virtual_keys.anthropic.virtual_keys
}

# Substring match on the virtual key name.
data "fastly_ai_runtime_control_virtual_keys" "chatbot" {
  search = "chatbot"
}

output "fastly_ai_runtime_control_chatbot_key_ids" {
  value = [for key in data.fastly_ai_runtime_control_virtual_keys.chatbot.virtual_keys : key.id]
}
