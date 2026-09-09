data "fastly_ai_runtime_control_sessions" "recent" {
  key  = fastly_ai_runtime_control_virtual_key.demo.id
  from = "2026-05-05T00:00:00Z"
}

# The `logs` attribute contains the raw prompts and completions exchanged with
# the AI provider, so avoid exposing it via an output.
output "fastly_ai_runtime_control_session_token_usage" {
  value = [
    for session in data.fastly_ai_runtime_control_sessions.recent.sessions : {
      id            = session.id
      model         = session.model
      input_tokens  = session.input_tokens
      output_tokens = session.output_tokens
    }
  ]
}
