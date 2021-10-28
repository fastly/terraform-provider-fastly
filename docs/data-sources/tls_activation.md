---
layout: "fastly"
page_title: "Fastly: fastly_tls_activation"
sidebar_current: "docs-fastly-datasource-tls_activation"
description: |-
Get information on Fastly TLS Activation.
---

# fastly_tls_activation

Use this data source to get information on a TLS activation, including the certificate used, and the domain on which TLS was enabled.

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination
of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination
with any of the others.

~> **Note:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single key.

## Example Usage

```terraform
data "fastly_tls_activation" "example" {
  domain = "example.com"
}
```