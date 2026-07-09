---
page_title: "fastly_acls Data Source - fastly"
subcategory: ""
description: |-
  Use this data source to retrieve a list of Fastly ACLs.
---

# fastly_acls (Data Source)

Use this data source to retrieve a list of [Fastly ACLs](https://www.fastly.com/documentation/reference/api/compute-acls/).

## Example Usage

```terraform
data "fastly_acls" "example" {}

output "fastly_acls_all" {
  value = data.fastly_acls.example.acls
}

output "fastly_acls_filtered" {
  # Example: get the ID of the ACL named "my_acl"
  value = one([
    for acl in data.fastly_acls.example.acls :
    acl.id if acl.name == "my_acl"
  ])
}
```

## Schema

### Read-Only

- `acls` (Attributes Set) List of all ACLs. (see [below for nested schema](#nestedatt--acls))
- `id` (String) Terraform data source identifier.

<a id="nestedatt--acls"></a>
### Nested Schema for `acls`

Read-Only:

- `id` (String) Identifier of the ACL.
- `name` (String) Name of the ACL.
