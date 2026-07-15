---
page_title: "fastly_acl_entries Resource - fastly"
subcategory: ""
description: |-
  Manages CIDR-based allow/block entries within a Fastly ACL.
---

# fastly_acl_entries (Resource)

Manages CIDR-based allow/block entries within a Fastly ACL.

By default, Terraform does not continue to manage the entries after the initial `terraform apply`. This allows you to make changes to ACL entries outside of Terraform using the [Fastly API](https://developer.fastly.com/reference/api/) or [Fastly CLI](https://developer.fastly.com/learning/tools/cli/) without Terraform resetting them.

To have Terraform continue managing the entries after creation (e.g., deleting any entries not defined in the config), set `manage_entries = true`.

~> **Note:** Use `manage_entries = true` cautiously. Terraform will overwrite external changes and delete any unmanaged entries.

## Example Usage

Basic usage (with seeded values, unmanaged after initial apply):

```terraform
resource "fastly_acl" "example" {
  name = "my_acl"
}

resource "fastly_acl_entries" "example" {
  acl_id = fastly_acl.example.id
  entries = {
    "192.0.2.0/24"    = "ALLOW"
    "198.51.100.0/24" = "BLOCK"
  }
}
```

Terraform-managed usage (where Terraform controls entries long-term):

```terraform
resource "fastly_acl" "example" {
  name = "my_acl"
}

resource "fastly_acl_entries" "example" {
  acl_id = fastly_acl.example.id
  entries = {
    "203.0.113.0/24"  = "BLOCK"
    "198.51.100.0/24" = "ALLOW"
  }
  manage_entries = true
}
```

## Schema

### Required

- `acl_id` (String) The ID of the ACL that the entries belong to.
- `entries` (Map of String) A map representing the entries in the ACL, where the keys are CIDR prefixes and the values are actions (`ALLOW` or `BLOCK`).

### Optional

- `manage_entries` (Boolean) Manage the ACL entries in Terraform (default: `false`). If `true`, Terraform will ensure that the ACL's entries match the entries in the Terraform configuration. When importing this resource, `manage_entries` is always set to `true`, so any ACL entries not present in the Terraform configuration will be deleted on the next apply.

### Read-Only

- `id` (String) Terraform resource identifier. Format: `acl_id/entries`.

## Import

Fastly ACL entries can be imported using the format `<acl_id>/entries`, e.g.

```shell
terraform import fastly_acl_entries.example <acl_id>/entries
```

Import reads the ACL's current entries from the Fastly API into state as-is; it does not modify the ACL or delete anything by itself.

However, `manage_entries` is always set to `true` on import, regardless of the value in your configuration. Because of this, your `entries` map must match what was imported before you run `terraform apply` — any entry present in state (from the import) but missing from your configuration will be deleted on the next apply. Run `terraform show` after importing to see the entries that were read in, and copy them into your configuration before applying.
