resource "fastly_ngwaf_account_list" "example" {
  name        = "shared-bot-ip-list"
  description = "Account list of known bot IPs shared across workspaces"
  type        = "ip"
  entries     = [
    "1.2.3.4",
    "5.6.7.8",
    "203.0.113.42"
  ]
}

# Example usage with a rule. 
resource "fastly_ngwaf_workspace_rule" "example" {
  workspace_id = fastly_ngwaf_workspace.example.id
  type = "request"
  description = "Example usage of a workspace list rule"
  enabled = true
  request_logging = "sampled"
  
  condition {
    field = "ip"
    operator = "not_in_list"
    # *********************************************
    # Account lists must be prefixed with 'corp.'
    # *********************************************
    value = "corp.${fastly_ngwaf_account_list.example.name}"
  }

  action {
    type = "block"
  }

  depends_on = [ fastly_ngwaf_account_list.example ]
}
