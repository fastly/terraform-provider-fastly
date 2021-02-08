---
layout: "fastly"
page_title: "Fastly: fastly_tls_private_key_ids"
sidebar_current: "docs-fastly-datasource-tls_private_key_ids"
description: |-
Get the list of TLS private key identifiers in Fastly.
---

# fastly_tls_private_key_ids

Use this data source to get the list of TLS private key identifiers in Fastly.

## Example Usage

```hcl
data "fastly_tls_private_key_ids" "demo" {}

data "fastly_tls_private_key" "example" {
  id = fastly_tls_private_key_ids.demo.ids[0]
}
```

## Arguments Reference

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `ids` - List of IDs of the TLS private keys
