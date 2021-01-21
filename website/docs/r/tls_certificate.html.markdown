---
layout: "fastly"
page_title: "Fastly: tls_certificate"
sidebar_current: "docs-fastly-resource-tls_certificate"
description: |-
Uploads a custom TLS certificate
---

# fastly_tls_certificate

Uploads a custom TLS certificate to Fastly to be used to terminate TLS traffic.

-> Each TLS certificate **must** have its corresponding private uploaded _prior_ to uploading the certificate. This can
be achieved in Terraform using [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html)

## Example Usage

Basic usage:

```hcl
resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "cert" {
  key_algorithm   = tls_private_key.key.algorithm
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example.com"
  }

  is_ca_certificate     = true
  validity_period_hours = 360

  allowed_uses = [
    "cert_signing",
    "server_auth",
  ]

  dns_names = ["example.com"]
}

resource "fastly_tls_private_key" "key" {
  key_pem = tls_private_key.key.private_key_pem
  name    = "tf-demo"
}

resource "fastly_tls_certificate" "example" {
  name = "tf-demo"
  certificate_body = tls_self_signed_cert.cert.cert_pem
  depends_on = [fastly_tls_private_key.key] // The private key has to be present before the certificate can be uploaded
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Human-readable name used to identify the certificate. Defaults to the certificate's Common Name or
first Subject Alternative Name entry
* `certificate_body` - (Required) PEM-formatted certificate

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `created_at` - Timestamp (GMT) when the certificate was created.
* `updated_at` - Timestamp (GMT) when the certificate was last updated.
* `issued_to` - The hostname for which a certificate was issued
* `issuer` - The certificate authority that issued the certificate
* `replace` - A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.
* `serial_number` - A value assigned by the issuer that is unique to a certificate.
* `signature_algorithm` - The algorithm used to sign the certificate.
* `domains` - All the domains (including wildcard domains) that are listed in any certificate's Subject Alternative
  Names (SAN) list.


## Import

A certificate can be imported using its Fastly certificate ID, e.g.

```
$ terraform import fastly_tls_certificate.demo xxxxxxxxxxx
```
