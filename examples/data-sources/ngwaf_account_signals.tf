data "fastly_ngwaf_account_signals" "account_signals" {}

output "fastly_ngwaf_account_signals_all" {
  value = data.fastly_ngwaf_account_signals.account_signals
}
