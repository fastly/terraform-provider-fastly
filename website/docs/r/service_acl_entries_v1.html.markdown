---
layout: "fastly"
page_title: "Fastly: service_acl_entries_v1"
sidebar_current: "docs-fastly-resource-service-v1"
description: |-
  Defines a set of Fastly ACL entries that can be used to populate a service ACL. 
---

# fastly_service_acl_entries_v1

Defines a set of Fastly ACL entries that can be used to populate a service ACL.  This resource will populate an ACL with the entries and will track their state.

~> **Warning:** Terraform will take precedence over any changes you make in the UI or API. Such changes are likely to be reversed if you run terraform again.  

If Terraform is being used to populate the initial content of an ACL which you intend to manage via API or UI, then the lifecycle `ignore_changes` field can be used with the resource.  An example of this configuration is provided below.    


## Example Usage

Basic usage:

```hcl
variable "myacl_name" {
	type = string
	default = "My ACL"
}

resource "fastly_service_v1" "myservice" {
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

resource "fastly_service_acl_entries_v1" "entries" {
  service_id = fastly_service_v1.myservice.id
  acl_id = {for d in fastly_service_v1.myservice.acl : d.name => d.acl_id}[var.myacl_name]
  entry {
    ip = "127.0.0.1"
    subnet = "24"
    negated = false
    comment = "ALC Entry 1"
  }
}
```

### Supporting API and UI ACL updates with ignore_changes

The following example demonstrates how the lifecycle ignore_change field can be used to suppress updates against the 
entries in an ACL.  If, after your first deploy, the Fastly API or UI is to be used to manage entries in an ACL, then this will stop Terraform realigning the remote state with the initial set of ACL entries defined in your HCL.

```hcl
...

resource "fastly_service_acl_entries_v1" "entries" {
  service_id = fastly_service_v1.myservice.id
  acl_id = {for d in fastly_service_v1.myservice.acl : d.name => d.acl_id}[var.myacl_name]
  entry {
    ip = "127.0.0.1"
    subnet = "24"
    negated = false
    comment = "ALC Entry 1"
  }
  
  lifecycle {
    ignore_changes = [entry,]
  }
  
}
```


## Argument Reference

The following arguments are supported:

* `service_id` - (Required) The ID of the Service that the acl belongs to
* `acl_id` - (Required) The ID of the acl that the items belong to
* `entry` - (Optional) A Set ACL entries that are applied to the service. Defined below

The `entry` block supports:

* `ip` - (Required, string) An IP address that is the focus for the ACL
* `subnet` - (Optional, string) An optional subnet mask applied to the IP address
* `negated` - (Optional, boolean) A boolean that will negate the match if true
* `comment` - (Optional, string) A personal freeform descriptive note



## Attributes Reference

* [fastly-acl](https://docs.fastly.com/api/config#acl)
* [fastly-acl_entry](https://docs.fastly.com/api/config#acl_entry)

## Import

This is an example of the import command being applied to the resource named `fastly_service_acl_entries_v1.entries`
The resource ID is a combined value of the `service_id` and `acl_id` separated by a forward slash.

```
$ terraform import fastly_service_acl_entries_v1.entries xxxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxx
```

If Terraform is already managing remote acl entries against a resource being imported then the user will be asked to remove it from the existing Terraform state.  
The following is an example of the Terraform state command to remove the resource named `fastly_service_acl_entries_v1.entries` from the Terraform state file.

```
$ terraform state rm fastly_service_acl_entries_v1.entries
``` 