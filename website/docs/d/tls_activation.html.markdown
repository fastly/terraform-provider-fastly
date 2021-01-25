---
layout: "fastly"
page_title: "Fastly: fastly_tls_activation"
sidebar_current: "docs-fastly-datasource-tls_activation"
description: |-
Get information on Fastly TLS Activation.
---

# fastly_tls_activation

Use this data source to get information on a TLS activation, including the certificate used, and the domain on which TLS was enabled.

## Example Usage

```hcl
data "fastly_tls_activation" "example" {
  domain = "example.com"
}
```

## Argument Reference

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination
of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination
with any of the others.

* `id` - (Optional) Fastly Activation ID. Conflicts with all other other filters.
* `certificate_id` - (Optional) ID of the TLS Certificate used
* `configuration_id` - (Optional) ID of the TLS Configuration used.
* `domain` - (Optional) Domain that TLS was enabled on.

~> **Note:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single key.

## Attribute Reference

In addition to the arguments specified above, the following attributes are also supported:

* `created_at` - Timestamp (GMT) when TLS was activated.
