---
layout: "fastly"
page_title: "Fastly: tls_platform_certificate"
sidebar_current: "docs-fastly-resource-tls_platform_certificate"
description: |-
Uploads a TLS certificate to the Platform TLS service
---

# fastly_tls_platform_certificate

Uploads a TLS certificate to the Fastly Platform TLS service.

-> Each TLS certificate **must** have its corresponding private key uploaded _prior_ to uploading the certificate. This
can be achieved in Terraform using [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html)

## Example Usage

Basic usage with self-signed CA:

```terraform
resource "tls_private_key" "ca_key" {
  algorithm = "RSA"
}

resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm   = tls_private_key.ca_key.algorithm
  private_key_pem = tls_private_key.ca_key.private_key_pem

  subject {
    common_name = "Example CA"
  }

  is_ca_certificate     = true
  validity_period_hours = 360

  allowed_uses = [
    "cert_signing",
    "server_auth",
  ]
}

resource "tls_cert_request" "example" {
  key_algorithm   = tls_private_key.key.algorithm
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example.com"
  }

  dns_names = ["example.com", "www.example.com"]
}

resource "tls_locally_signed_cert" "cert" {
  cert_request_pem   = tls_cert_request.example.cert_request_pem
  ca_key_algorithm   = tls_private_key.ca_key.algorithm
  ca_private_key_pem = tls_private_key.ca_key.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 360

  allowed_uses = [
    "cert_signing",
    "server_auth",
  ]
}

data "fastly_tls_configuration" "config" {
  tls_service = "PLATFORM"
}

resource "fastly_tls_private_key" "key" {
  key_pem = tls_private_key.key.private_key_pem
  name    = "tf-demo"
}

resource "fastly_tls_platform_certificate" "cert" {
  certificate_body   = tls_locally_signed_cert.cert.cert_pem
  intermediates_blob = tls_self_signed_cert.ca.cert_pem

  configuration_id     = data.fastly_tls_configuration.config.id
  allow_untrusted_root = true

  depends_on = [fastly_tls_private_key.key]
}
```

## Import

A certificate can be imported using its Fastly certificate ID, e.g.

```txt
$ terraform import fastly_tls_platform_certificate.demo xxxxxxxxxxx
```