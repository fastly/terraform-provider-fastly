# IMPORTANT: Deleting a TSIG Key requires first removing the inbound_tsig_key_id reference in any DNS Zone that uses it.
# This requires a two-step `terraform apply` as we can't guarantee deletion order.
# e.g. removal of inbound_tsig_key_id within fastly_dns_zone might not finish first.
resource "fastly_tsig_key" "example" {
  name        = "example.com."
  algorithm   = "hmac-sha256"
  secret      = "c2VjcmV0a2V5MTIzNDU2Nzg="
  description = "My TSIG key"
}
