resource "fastly_dns_zone" "example" {
    name        = "example.com."
    description = "My secondary DNS zone"

    xfr_config_inbound {
      inbound_tsig_key_id = fastly_tsig_key.example.id

      primaries {
        address     = "192.0.2.1"
        description = "Primary nameserver #1"
      }

      primaries {
        address     = "192.0.2.2"
        description = "Primary nameserver #1"
      }
    }
  }