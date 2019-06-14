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

TODO Provide an example defining a basic dictionary items.


```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) The ID of the Service that the dictionary belongs to
* `dictionary_id` - (Required) The ID of the dictionary that the items belong to
* `items` - (Optional) A Map representing an entry in the dictionary, (key/value)


## Attributes Reference

[fastly-dictionary]: https://docs.fastly.com/api/config#dictionary
[fastly-dictionary_item]: https://docs.fastly.com/api/config#dictionary_item
