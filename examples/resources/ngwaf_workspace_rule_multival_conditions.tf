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
  description     = "Block requests with specific header patterns"
  enabled         = true
  request_logging = "sampled"
  group_operator  = "all"

  action {
    type = "block"
  }

  multival_condition {
    field          = "request_header"
    operator       = "exists"
    group_operator = "any"

    condition {
      field    = "name"
      operator = "does_not_equal"
      value    = "Header-Sample"
    }

    condition {
      field    = "name"
      operator = "contains"
      value    = "X-API-Key"
    }
  }
}
