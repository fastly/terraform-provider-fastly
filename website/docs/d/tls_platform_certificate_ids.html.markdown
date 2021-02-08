---
layout: "fastly"
page_title: "Fastly: fastly_tls_platform_certificate_ids"
sidebar_current: "docs-fastly-datasource-tls_platform_certificate_ids"
description: |-
Get IDs of available Platform TLS certificates.
---

# fastly_tls_platform_certificate_ids

Use this data source to get the IDs of available Platform TLS Certificates for use with other resources.

## Example Usage

```hcl
data "fastly_tls_platform_certificate_ids" "example" {}

data "fastly_tls_platform_certificate" "example" {
  id = data.fastly_tls_platform_certificate_ids.example.ids[0]
}
```

## Argument Reference

This data source has no arguments

## Attribute Reference

* `ids` - List of IDs corresponding to Platform TLS certificates