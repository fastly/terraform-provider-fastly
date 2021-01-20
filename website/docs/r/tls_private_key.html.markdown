---
layout: "fastly"
page_title: "Fastly: tls_private_key"
sidebar_current: "docs-fastly-resource-tls_private_key"
description: |-
Uploads a Custom TLS Private Key
---

# fastly_tls_private_key

Uploads a Custom TLS Private Key to Fastly. This can be combined with a `fastly_tls_custom_certificate` resource to provide a TLS Certificate able to be applied to a Fastly Service.

The Private Key resource requires a key in PEM format, and a name to identify it.

## Example Usage

Basic usage:

```hcl
resource "tls_private_key" "demo" {
  algorithm = "RSA"
}

resource "fastly_tls_private_key" "demo" {
  key_pem = tls_private_key.demo.private_key_pem
  name    = "tf-demo"
}
```

## Argument Reference

The following arguments are supported:

* `key_pem` - (Required) Private key in PEM format. Not readable after uploading.
* `name` - (Required) Human-readable name used to identify the private key.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `created_at` - Timestamp (GMT) when the private key was created.
* `key_length` - The key length used to generate the private key.
* `key_type` - The algorithm used to generate the private key. Must be RSA.
* `replace` - A boolean indicating whether Fastly recommends replacing this private key.
* `public_key_sha1` - A hash of the associated public key useful for safely identifying the key after it has been uploaded.

## Import

A Private Key can be imported using its ID, e.g.

```
$ terraform import fastly_tls_private_key.demo xxxxxxxxxxx
```
