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
  description     = ""
  enabled         = true
  group_operator  = "all"

  condition {
    field    = "method"
    operator = "equals"
    value    = "POST"
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/login"
  }

  action {
    type   = "templated_signal"
    signal = "LOGINATTEMPT"
  }
}
