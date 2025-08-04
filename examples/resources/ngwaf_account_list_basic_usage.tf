resource "fastly_ngwaf_account_list" "example" {
  name        = "shared-bot-ip-list"
  description = "List of known bot IPs shared across workspaces"
  type        = "ip"
  entries     = [
    "1.2.3.4",
    "5.6.7.8",
    "203.0.113.42"
  ]
}
