data "fastly_dns_zones" "example" {}

output "fastly_dns_zones_all" {
  value = data.fastly_dns_zones.example.zones
}
