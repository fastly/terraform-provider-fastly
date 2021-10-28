---
layout: "fastly"
page_title: "Fastly: fastly_tls_certificate"
sidebar_current: "docs-fastly-datasource-tls_certificate"
description: |-
Get information on Fastly TLS certificate.
---

# fastly_tls_certificate

Use this data source to get information of a TLS certificate for use with other resources.

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination
of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination
with any of the others.

~> **Note:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single key.

## Example Usage

```terraform
data "fastly_tls_certificate" "example" {
  name = "example.com"
}
```