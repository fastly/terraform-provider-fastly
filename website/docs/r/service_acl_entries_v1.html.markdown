---
layout: "fastly"
page_title: "Fastly: service_acl_entries_v1"
sidebar_current: "docs-fastly-resource-service-v1"
description: |-
  Provides a grouping of fastly acl entries that can be applied to a service. 
---

# fastly_service_acl_entries_v1

Provides a grouping of fastly acl entries that can be applied to a service.

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

[fastly-acl]: https://docs.fastly.com/api/config#acl
[fastly-acl_entry]: https://docs.fastly.com/api/config#acl_entry