---
layout: "fastly"
page_title: "Fastly: fastly_tls_certificate"
sidebar_current: "docs-fastly-datasource-tls_certificate"
description: |-
Get information on Fastly TLS certificate.
---

# fastly_tls_certificate

Use this data source to get information from a TLS certificate for use with other resources.

## Example Usage

```hcl
data "fastly_tls_certificate" "example" {
  name = "example.com"
}
```

## Argument Reference

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination with any of the others.

* `id` - (Optional) Unique ID assigned to certificate by Fastly. Conflicts with all the other filters.
* `name` - (Optional) Human-readable name used to identify the certificate
* `issued_to` - (Optional) The hostname for which a certificate was issued
* `domains` - (Optional) Domains that are listed in any certificate's Subject Alternative Names (SAN) list
* `issuer` - (Optional) The certificate authority that issued the certificate

## Attribute Reference

* `created_at` - Timestamp (GMT) when the certificate was created
* `updated_at` - Timestamp (GMT) when the certificate was last updated
* `replace` - A recommendation from Fastly indicating the key associated with this certificate is in need of rotation
* `serial_number` - A value assigned by the issuer that is unique to a certificate
* `signature_algorithm` - The algorithm used to sign the certificate
