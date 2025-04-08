data "fastly_acls" "fastly_all" {}

data "fastly_acls" "fastly_by_name" {
    acls {
        name = "My ACL"
    }
}

output "fastly_acls_all" {
  value = data.fastly_acls.fastly_all
}
output "fastly_acls_filtered" {
  value = data.fastly_acls.fastly_by_name
}
