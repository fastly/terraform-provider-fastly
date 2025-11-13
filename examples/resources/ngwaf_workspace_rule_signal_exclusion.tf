resource "fastly_ngwaf_workspace" "example" {
  name                            = "example"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_workspace_rule" "exclude_xss_signal" {
  workspace_id    = fastly_ngwaf_workspace.example.id
  type            = "signal"
  description     = "Exclude XSS signal to address a false positive"
  enabled         = true
  group_operator  = "all"

  condition {
    field    = "path"
    operator = "like"
    value    = "/contact-form"
  }
  action {
    type   = "exclude_signal"
    signal = "XSS"
  }
}