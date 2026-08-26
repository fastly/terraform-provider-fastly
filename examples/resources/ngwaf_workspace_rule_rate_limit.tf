resource "fastly_ngwaf_workspace" "example" {
  name                            = "example"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {}
}


resource "fastly_ngwaf_workspace_signal" "demo_signal" {
  workspace_id = fastly_ngwaf_workspace.example.id
  name         = "demo"
  description  = "A description of my signal."
}

resource "fastly_ngwaf_workspace_rule" "ip_limit" {
  workspace_id    = fastly_ngwaf_workspace.example.id
  type            = "rate_limit"
  description     = "Rate limit demo rule-updated"
  enabled         = true

  condition {
    field = "ip"
    operator = "equals"
    value = "1.2.3.4"
  }

  rate_limit {
    signal    = "site.demo"
    threshold = 100
    interval  = 60
    duration  = 300

    client_identifiers {
      type = "ip"
    }
  }

  action {
    signal = "SUSPECTED-BOT"
    type = "block_signal"
  }
}

