---
layout: "fastly"
page_title: "Fastly: fastly_tls_certificate_ids"
sidebar_current: "docs-fastly-datasource-tls_certificate_ids"
description: |-
Get IDs of available TLS certificates.
---

# fastly_tls_certificate_ids

Use this data source to get the IDs of available TLS configurations for use with other resources.

## Example Usage

```hcl
data "fastly_tls_certificate_ids" "example" {}

resource "fastly_tls_activation" "example" {
  certificate_id = data.fastly_tls_certificate_ids.example.ids[0]
  // ...
}
```

## Argument Reference

This data source has no arguments

## Attribute Reference

* `ids` - List of IDs corresponding to TLS certificates
