  data "fastly_domains_v1" "example" {
  }

   output "all_domains" {
    value = data.fastly_domains_v1.example.domains
  }

  output "total_domains" {
    value = data.fastly_domains_v1.example.total
    
  }