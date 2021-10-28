---
layout: "fastly"
page_title: "Fastly: fastly_tls_domain"
sidebar_current: "docs-fastly-datasource-tls_domain"
description: |-
Get IDs of activations, certificates and subscriptions associated with a domain.
---

# fastly_tls_domain

Use this data source to get the IDs of activations, certificates and subscriptions associated with a domain.

## Example Usage

```terraform
data "fastly_tls_domain" "domain" {
  domain = "example.com"
}
```