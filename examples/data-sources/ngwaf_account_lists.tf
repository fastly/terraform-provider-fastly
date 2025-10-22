data "fastly_ngwaf_account_lists" "account_lists" {}

output "fastly_ngwaf_account_lists_all" {
  value = data.fastly_ngwaf_account_lists.account_lists
}
