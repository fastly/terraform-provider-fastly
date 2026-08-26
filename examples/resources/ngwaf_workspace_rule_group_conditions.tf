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
  description     = "Block requests with grouped conditions"
  enabled         = true
  request_logging = "sampled"
  group_operator  = "all"

  action {
    type = "block"
  }

  # This group uses "any" - matches if 'any' condition is true
  group_condition {
    group_operator = "any"

    condition {
      field    = "protocol_version"
      operator = "equals"
      value    = "HTTP/1.0"
    }

    condition {
      field    = "method"
      operator = "equals"
      value    = "HEAD"
    }

    condition {
      field    = "domain"
      operator = "equals"
      value    = "example.com"
    }
  }

  # This group uses "all" - matches only if 'all' conditions are true
  group_condition {
    group_operator = "all"

    condition {
      field    = "country"
      operator = "equals"
      value    = "AD"
    }

    condition {
      field    = "method"
      operator = "equals"
      value    = "POST"
    }
  }
}
