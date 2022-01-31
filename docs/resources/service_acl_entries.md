---
layout: "fastly"
page_title: "Fastly: service_acl_entries"
sidebar_current: "docs-fastly-resource-service-acl-entries"
description: |-
  Defines a set of Fastly ACL entries that can be used to populate a service ACL.
---

# fastly_service_acl_entries

Defines a set of Fastly ACL entries that can be used to populate a service ACL.  This resource will populate an ACL with the entries and will track their state.

~> **Warning:** Terraform will take precedence over any changes you make in the UI or API. Such changes are likely to be reversed if you run Terraform again.

~> **Note:** By default the Terraform provider allows you to externally manage the entries via API or UI.
If you wish to apply your changes in the HCL, then you should explicitly set the `manage_entries` attribute. An example of this configuration is provided below.

## Example Usage (Terraform >= 0.12.6)

Basic usage:

```terraform
variable "myacl_name" {
  type = string
  default = "My ACL"
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name = "AWS S3 hosting"
    port = 80
  }

  acl {
    name = var.myacl_name
  }

  force_destroy = true
}

resource "fastly_service_acl_entries" "entries" {
  for_each = {
  for d in fastly_service_vcl.myservice.acl : d.name => d if d.name == var.myacl_name
  }
  service_id = fastly_service_vcl.myservice.id
  acl_id = each.value.acl_id
  entry {
    ip = "127.0.0.1"
    subnet = "24"
    negated = false
    comment = "ACL Entry 1"
  }
}
```

Complex object usage:

The following example demonstrates the use of dynamic nested blocks to create ACL entries.

```terraform
locals {
  acl_name = "my_acl"
  acl_entries = [
    {
      ip      = "1.2.3.4"
      comment = "acl_entry_1"
    },
    {
      ip      = "1.2.3.5"
      comment = "acl_entry_2"
    },
    {
      ip      = "1.2.3.6"
      comment = "acl_entry_3"
    }
  ]
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "1.2.3.4"
    name    = "localhost"
    port    = 80
  }

  acl {
    name = local.acl_name
  }

  force_destroy = true
}

resource "fastly_service_acl_entries" "entries" {
  for_each = {
  for d in fastly_service_vcl.myservice.acl : d.name => d if d.name == local.acl_name
  }
  service_id = fastly_service_vcl.myservice.id
  acl_id = each.value.acl_id
  dynamic "entry" {
    for_each = [for e in local.acl_entries : {
      ip      = e.ip
      comment = e.comment
    }]

    content {
      ip      = entry.value.ip
      subnet  = 22
      comment = entry.value.comment
      negated = false
    }
  }
}
```

## Example Usage (Terraform >= 0.12.0 && < 0.12.6)

`for_each` attributes were not available in Terraform before 0.12.6, however, users can still use `for` expressions to achieve
similar behaviour as seen in the example below.

~> **Warning:** Terraform might not properly calculate implicit dependencies on computed attributes when using `for` expressions

For scenarios such as adding an ACL to a service and at the same time, creating the ACL entries (`fastly_service_acl_entries`)
resource, Terraform will not calculate implicit dependencies correctly on `for` expressions. This will result in index lookup
problems and the execution will fail.

For those scenarios, it's recommended to split the changes into two distinct steps:

1. Add the `acl` block to the `fastly_service_vcl` and apply the changes
2. Add the `fastly_service_acl_entries` resource with the `for` expressions to the HCL and apply the changes

Usage:

```terraform
variable "myacl_name" {
  type    = string
  default = "My ACL"
}

resource "fastly_service_vcl" "myservice" {
  #...
  acl {
    name = var.myacl_name
  }
  #...
}

resource "fastly_service_acl_entries" "entries" {
  service_id = fastly_service_vcl.myservice.id
  acl_id     = {for d in fastly_service_vcl.myservice.acl : d.name => d.acl_id}[var.myacl_name]
  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "ACL Entry 1"
  }
}
```

### Reapplying original entries with `managed_entries` if the state of the entries drifts

By default the user is opted out from reapplying the original changes if the entries are managed externally.
The following example demonstrates how the `manage_entries` field can be used to reapply the changes defined in the HCL if the state of the entries drifts.
When the value is explicitly set to 'true', Terraform will keep the original changes and discard any other changes made under this resource outside of Terraform.

~> **Warning:** You will lose externally managed entries if `manage_entries=true`.

~> **Note:** The `ignore_changes` built-in meta-argument takes precedence over `manage_entries` regardless of its value.

```terraform
#...

resource "fastly_service_acl_entries" "entries" {
  for_each = {
    for d in fastly_service_vcl.myservice.acl : d.name => d if d.name == var.myacl_name
  }
  service_id     = fastly_service_vcl.myservice.id
  acl_id         = each.value.acl_id
  manage_entries = true
  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "ACL Entry 1"
  }
}
```

## Attributes Reference

* [fastly-acl](https://developer.fastly.com/reference/api/acls/acl/)
* [fastly-acl_entry](https://developer.fastly.com/reference/api/acls/acl-entry/)

## Import

This is an example of the import command being applied to the resource named `fastly_service_acl_entries.entries`
The resource ID is a combined value of the `service_id` and `acl_id` separated by a forward slash.

```sh
$ terraform import fastly_service_acl_entries.entries xxxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxx
```

If Terraform is already managing remote acl entries against a resource being imported then the user will be asked to remove it from the existing Terraform state.
The following is an example of the Terraform state command to remove the resource named `fastly_service_acl_entries.entries` from the Terraform state file.

```sh
$ terraform state rm fastly_service_acl_entries.entries
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **acl_id** (String) The ID of the ACL that the items belong to
- **service_id** (String) The ID of the Service that the ACL belongs to

### Optional

- **entry** (Block Set, Max: 10000) ACL Entries (see [below for nested schema](#nestedblock--entry))
- **id** (String) The ID of this resource.
- **manage_entries** (Boolean) Whether to reapply changes if the state of the entries drifts, i.e. if entries are managed externally

<a id="nestedblock--entry"></a>
### Nested Schema for `entry`

Required:

- **ip** (String) An IP address that is the focus for the ACL

Optional:

- **comment** (String) A personal freeform descriptive note
- **negated** (Boolean) A boolean that will negate the match if true
- **subnet** (String) An optional subnet mask applied to the IP address

Read-Only:

- **id** (String) The unique ID of the entry