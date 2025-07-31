resource "fastly_ngwaf_account_signal" "example" {
  applies_to       = ["*"]
  description      = "example"
  name             = "Test Name"
}
