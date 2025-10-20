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
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "request"
  description      = "example"
  enabled          = true
  request_logging  = "sampled"
  group_operator   = "all"

  action {
    type = "block"
  }

  condition {
    field    = "ip"
    operator = "equals"
    value    = "127.0.0.1"
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/login"
  }

  condition {
    field    = "agent_name"
    operator = "equals"
    value    = "host-001"
  }

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
}

resource "fastly_ngwaf_workspace_rule" "LOGINATTEMPT" {
  workspace_id   = fastly_ngwaf_workspace.waf.id
  type           = "templated_signal"
  enabled        = true
  group_operator = "all"
  description    = ""

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
