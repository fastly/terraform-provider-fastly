  resource "fastly_ngwaf_thresholds" "demo" {
    action       = "block"
    dont_notify  = false
    duration     = 86400
    enabled      = true
    interval     = 3600
    limit        = 10
    name         = "%s"
    signal       = "sqli"
    workspace_id = fastly_ngwaf_workspace.example.id
  }