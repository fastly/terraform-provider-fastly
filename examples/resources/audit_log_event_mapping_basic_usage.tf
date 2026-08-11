resource "fastly_integration" "example" {
  name        = "Example Slack integration"
  description = "Notifies #alerts on Slack"
  type        = "slack"

  config = {
    webhook = "https://hooks.slack.com/services/xxx/xxx/xxx"
  }
}

resource "fastly_audit_log_event_mapping" "example" {
  name            = "Example mapping"
  description     = "Sends a notification when any user logs in"
  scope_type      = "account"
  event_types     = ["user.login"]
  integration_ids = [fastly_integration.example.id]
}
