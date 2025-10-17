data "fastly_ngwaf_account_rules" "account_rules" {}

output "fastly_ngwaf_account_rules_all" {
  value = data.fastly_ngwaf_account_rules.account_rules
}
