resource "fastly_ngwaf_workspace" "example" {
  name                            = "example"
  description                     = "Workspace with custom list"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 403

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_list" "example" {
  workspace_id = fastly_ngwaf_workspace.example.id
  name         = "local-allowlist"
  description  = "IP allowlist for this workspace"
  type         = "ip"
  entries      = [
    "192.168.0.1",
    "10.0.0.1"
  ]
}
