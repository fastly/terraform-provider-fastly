---
page_title: "fastly_acl Resource - fastly"
subcategory: ""
description: |-
  Provides an Access Control List (ACL) that defines CIDR-based access rules and is accessible to Compute services during request processing.
---

# fastly_acl (Resource)

Provides an Access Control List (ACL) that defines CIDR-based access rules (e.g., allow/block IP ranges) and is accessible to Compute services during request processing.

This resource is versionless: it is not tied to a service version and is managed independently of any `fastly_service_compute` resource.

## Example Usage

```terraform
resource "fastly_acl" "example" {
  name = "my_acl"
}
```

## Schema

### Required

- `name` (String) A unique name to identify the ACL. Changing this attribute will delete and recreate the ACL, discarding its current entries. Any `resource_link` referencing this ACL must be removed first.

### Read-Only

- `id` (String) The ID of this ACL.

## Import

Fastly ACLs can be imported using their ACL ID, e.g.

```shell
terraform import fastly_acl.example <acl_id>
```
