data "fastly_compute_acls" "example" {}

output "fastly_compute_acls_all" {
  value = data.fastly_compute_acls.example.acls
}

output "fastly_compute_acls_filtered" {
  # Example: get the ID of the ACL named "My ACL"
  value = one([
    for acl in data.fastly_compute_acls.example.acls :
    acl.id if acl.name == "My ACL"
  ])
}
