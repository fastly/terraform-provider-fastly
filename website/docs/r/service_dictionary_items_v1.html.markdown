---
layout: "fastly"
page_title: "Fastly: service_dictionary_items_v1"
sidebar_current: "docs-fastly-resource-service-v1"
description: |-
  Provides a grouping of fastly dictionary items that can be applied to a service. 
---

# fastly_service_dictionary_items_v1

Provides a grouping of fastly dictionary items that can be applied to a service.

## Example Usage

Basic usage:

```hcl

variable "mydict_name" {
	type = string
	default = "My Dictionary"
}

resource "fastly_service_v1" "myservice" {
  name = "demofastly"

  domain {
      name    = "demo.notexample.com"
      comment = "demo"
  }

  backend {
      address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
      name    = "AWS S3 hosting"
      port    = 80
    }

  dictionary {
	name       = var.mydict_name
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items_v1" "items" {
    service_id = "${fastly_service_v1.myservice.id}"
    dictionary_id = "${{for s in fastly_service_v1.myservice.dictionary : s.name => s.dictionary_id}[var.mydict_name]}"
    items = {
        key1: "value1"
        key2: "value2"
    }
}

```

Complex object usage:

```hcl

variable "mydict" {
  type = object({ name=string, items=map(string) })
  default = {
    name = "My Dictionary"
    items = {
      key1: "value1x"
      key2: "value2x"
    }
  }
}

resource "fastly_service_v1" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dictionary {
    name = var.mydict.name
  }

  force_destroy = true
  }

resource "fastly_service_dictionary_items_v1" "items" {
  service_id = fastly_service_v1.myservice.id
  dictionary_id = {for d in fastly_service_v1.myservice.dictionary : d.name => d.dictionary_id}[var.mydict.name]
  items = var.mydict.items
}

```

Expression and functions usage:

```hcl
// Local variables used when formating values for the "My Project Dictionary" example
locals {
  dictionary_name = "My Project Dictionary"
  host_base = "demo.ocnotexample.com"
  host_divisions = ["alpha", "beta", "gamma", "delta"]
}

// Define the standard service that will be used to manage the dictionaries.
resource "fastly_service_v1" "myservice" {
  name = "demofastly"

  domain {
    name = "demo.ocnotexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.ocnotexample.com.s3-website-us-west-2.amazonaws.com"
    name = "AWS S3 hosting"
    port = 80
  }

  dictionary {
    name = local.dictionary_name
  }

  force_destroy = true
}

// This resource is dynamically creating the items from the local variables through for expressions and functions.
resource "fastly_service_dictionary_items_v1" "project" {
  service_id = fastly_service_v1.myservice.id
  dictionary_id = {for d in fastly_service_v1.myservice.dictionary : d.name => d.dictionary_id}[local.dictionary_name]
  items = {
    for division in local.host_divisions:
      division => format("%s.%s", division, local.host_base)
  }

}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) The ID of the Service that the dictionary belongs to
* `dictionary_id` - (Required) The ID of the dictionary that the items belong to
* `items` - (Optional) A Map representing an entry in the dictionary, (key/value)


## Attributes Reference

[fastly-dictionary]: https://docs.fastly.com/api/config#dictionary
[fastly-dictionary_item]: https://docs.fastly.com/api/config#dictionary_item
