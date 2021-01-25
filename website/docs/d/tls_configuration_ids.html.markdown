---
layout: "fastly"
page_title: "Fastly: fastly_tls_configuration_ids"
sidebar_current: "docs-fastly-datasource-tls_configuration_ids"
description: |-
Get IDs of available TLS Configurations.
---

# fastly_tls_configuration_ids

Use this data source to get the IDs of available TLS configurations for use with other resources.

## Example Usage

```hcl
data "fastly_tls_configuration_ids" "example" {}

resource "fastly_tls_activation" "example" {
  configuration_id = data.fastly_tls_configuration.example.ids[0]
  // ...
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `ids` - List of IDs corresponding to TLS configurations