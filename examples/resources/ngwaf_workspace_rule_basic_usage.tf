resource "fastly_ngwaf_workspace" "example" {
  name                            = "example"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_workspace_rule" "example" {
  workspace_id    = fastly_ngwaf_workspace.example.id
  type            = "request"
  description     = "Block requests from specific IP to login path"
  enabled         = true
  request_logging = "sampled"
  group_operator  = "all"

  action {
    type = "block"
  }

  condition {
    field    = "ip"
    operator = "equals"
    value    = "192.0.2.1"
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/login"
  }
}
