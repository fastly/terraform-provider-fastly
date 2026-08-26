data "fastly_domains" "example" {
}

output "all_domains" {
  value = data.fastly_domains.example.domains
}

output "total_domains" {
  value = data.fastly_domains.example.total
}