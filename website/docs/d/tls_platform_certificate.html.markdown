---
layout: "fastly"
page_title: "Fastly: fastly_tls_platform_certificate"
sidebar_current: "docs-fastly-datasource-tls_platform_certificate"
description: |-
Get information on Fastly Platform TLS certificate.
---

# fastly_tls_platform_certificate

Use this data source to get information of a Platform TLS certificate for use with other resources.

## Example Usage

```hcl
data "fastly_tls_platform_certificate" "example" {
  domains = ["example.com"]
}
```

## Argument Reference

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination with any of the others.

* `id` - (Optional) Unique ID assigned to certificate by Fastly. Conflicts with all the other filters.
* `domains` - (Optional) Domains that are listed in any certificate's Subject Alternative Names (SAN) list

## Attribute Reference

* `created_at` - Timestamp (GMT) when the certificate was created
* `updated_at` - Timestamp (GMT) when the certificate was last updated
* `not_after` - Timestamp (GMT) when the certificate will expire.
* `not_before` - Timestamp (GMT) when the certificate will become valid.
* `replace` - A recommendation from Fastly indicating the key associated with this certificate is in need of rotation
* `configuration_id` - ID of TLS configuration used to terminate TLS traffic.