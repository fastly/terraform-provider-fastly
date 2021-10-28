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

```terraform
resource "tls_private_key" "demo" {
  algorithm = "RSA"
}

resource "fastly_tls_private_key" "demo" {
  key_pem = tls_private_key.demo.private_key_pem
  name    = "tf-demo"
}
```

## Import

A Private Key can be imported using its ID, e.g.

```txt
$ terraform import fastly_tls_private_key.demo xxxxxxxxxxx
```