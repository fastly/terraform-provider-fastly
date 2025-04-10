---
page_title: "Fastly: fastly_compute_acl_entries"
subcategory: "Compute"
description: |-
  Provides a Fastly ACL Entries resource that can be used to add, update, and remove entries from a Compute ACL.
---

# fastly_compute_acl_entries

Provides a Fastly resource to add, update, and remove entries from a Compute ACL.

~> **Note:** This resource is distinct from `fastly_service_acl_entries` which manages ACL entries in the context of a Fastly service. This resource is for managing entries in Compute@Edge ACLs.

## Example Usage

```hcl
```terraform
resource "fastly_compute_acl" "my_acl" {
  name = "My ACL"
  force_destroy = true
}

resource "fastly_compute_acl_entries" "entries" {
  acl_id = fastly_acl.my_acl.acl_id
  force_destroy = true

  entry {
    prefix = "127.0.0.1/32"
    action = "BLOCK"
  }

  entry {
    ip = "192.168.0.1/32"
    action = "ALLOW"
  }
}
```
```

## Argument Reference

The following arguments are supported:

* `acl_id` - (Required) The ID of the ACL that the entries belong to.
* `entries` - (Optional) Set of ACL entries. See below for ACL entry properties.
* `force_destroy` - (Optional) Allow all ACL entries to be deleted during destroy. Default is `false`.
* `manage_entries` - (Optional) Boolean flag to control whether to reapply changes if the state of the entries drifts, i.e. if entries are managed externally. Default is `true`.

The `entries` attribute supports the following keys:

* `action` - (Required) The action to take on the entry. Valid values are `allow` or `block`.
* `prefix` - (Required) The ACL entry prefix in Classless Inter-Domain Routing (CIDR) notation.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the ACL.

## Import

Fastly ACL entries can be imported using the ACL ID, e.g.

```
terraform import fastly_compute_acl_entries.example 7d991f5f-7c40-4c8c-a0c1-6ea5e45e4bcf/entries
```