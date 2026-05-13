data "fastly_tsig_keys" "example" {}

output "fastly_tsig_keys_all" {
  value = data.fastly_tsig_keys.example.keys
}
