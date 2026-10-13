data "fastly_ai_runtime_control_provider_connections" "configured" {}

output "fastly_ai_runtime_control_provider_connections_all" {
  value = data.fastly_ai_runtime_control_provider_connections.configured.provider_connections
}

output "fastly_ai_runtime_control_openai_connection_id" {
  # get the ID of the connection named "OpenAI"
  value = one([
    for connection in data.fastly_ai_runtime_control_provider_connections.configured.provider_connections :
    connection.id if connection.name == "OpenAI"
  ])
}
