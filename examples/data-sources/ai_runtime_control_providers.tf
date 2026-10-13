data "fastly_ai_runtime_control_providers" "supported" {}

output "fastly_ai_runtime_control_providers_all" {
  value = data.fastly_ai_runtime_control_providers.supported.providers
}

output "fastly_ai_runtime_control_anthropic_models" {
  # get the model IDs offered by the Anthropic provider
  value = one([
    for provider in data.fastly_ai_runtime_control_providers.supported.providers :
    [for model in provider.models : model.id] if provider.id == "anthropic"
  ])
}
