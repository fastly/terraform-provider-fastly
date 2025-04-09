data "fastly_compute_acls" "fastly_all" {}

data "fastly_compute_acls" "fastly_by_name" {
    acls {
        name = "My ACL"
    }
}

output "fastly_compute_acls_all" {
  value = data.fastly_compute_acls.fastly_all
}
output "fastly_compute_acls_filtered" {
  value = data.fastly_compute_acls.fastly_by_name
}
