---
layout: "fastly"
page_title: "Fastly: service_dictionary_items_v1"
sidebar_current: "docs-fastly-resource-service-v1"
description: |-
  Provides a grouping of Fastly dictionary items that can be applied to a service. 
---

# fastly_service_dictionary_items_v1

Defines a map of Fastly dictionary items that can be used to populate a service dictionary.
 
This resource will populate a dictionary with the items and will track their state.
Terraform will take precedence over any changes you make in the UI or API. Such changes are likely to be reversed if you run terraform again.  

If Terraform is being used to populate the initial content of a dictionary which you intend to manage via API or UI, then the lifecycle `ignore_changes` field can be used with the resource.  An example of this configuration is provided below.    


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

###Supporting API and UI dictionary updates with ignore_changes

The following example demonstrates how the lifecycle ignore_change field can be used to suppress updates against the 
items in a dictionary.  If, after your first deploy, the Fastly API or UI is to be used to manage items in a dictionary, then this will stop Terraform realigning the remote state with the initial set of dictionary items defined in your HCL.

```hcl
...

resource "fastly_service_dictionary_items_v1" "items" {
    service_id = "${fastly_service_v1.myservice.id}"
    dictionary_id = "${{for s in fastly_service_v1.myservice.dictionary : s.name => s.dictionary_id}[var.mydict_name]}"
    items = {
        key1: "value1"
        key2: "value2"
    }
    
    lifecycle {
      ignore_changes = [items,]
    }
}


```


## Argument Reference

The following arguments are supported:

* `service_id` - (Required) The ID of the service that the dictionary belongs to
* `dictionary_id` - (Required) The ID of the dictionary that the items belong to
* `items` - (Optional) A map representing an entry in the dictionary, (key/value)


## Attributes Reference

* [fastly-dictionary](https://docs.fastly.com/api/config#dictionary)
* [fastly-dictionary_item](https://docs.fastly.com/api/config#dictionary_item)

## Import

This is an example of the import command being applied to the resource named `fastly_service_dictionary_items_v1.items`
The resource ID is a combined value of the `service_id` and `dictionary_id` separated by a forward slash.

```
$ terraform import fastly_service_dictionary_items_v1.items xxxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxx
```

If Terraform is already managing remote dictionary items against a resource being imported then the user will be asked to remove it from the existing Terraform state.  
The following is an example of the Terraform state command to remove the resource named `fastly_service_dictionary_items_v1.items` from the Terraform state file.

```
$ terraform state rm fastly_service_dictionary_items_v1.items
``` 
