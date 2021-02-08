---
layout: "fastly"
page_title: "Fastly: tls_activation"
sidebar_current: "docs-fastly-resource-tls_activation"
description: |-
Enables TLS on a domain
---

# fastly_tls_activation

Enables TLS on a domain using a specified custom TLS certificate.

~> **Note:** The Fastly service must be provisioned _prior_ to enabling TLS on it. This can be achieved in Terraform using [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html).

## Example Usage

Basic usage:

```hcl
resource "fastly_service_v1" "demo" {
  name = "my-service"

  domain {
    name    = "example.com"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_private_key" "demo" {
  key_pem = "..."
  name = "demo-key"
}

resource "fastly_tls_certificate" "demo" {
  certificate_body = "..."
  name = "demo-cert"
  depends_on = [fastly_tls_private_key.demo]
}

resource "fastly_tls_activation" "test" {
  certificate_id = fastly_tls_certificate.demo.id
  domain = "example.com"
  depends_on = [fastly_service_v1.demo]
}
```

## Argument Reference

The following arguments are supported:

* `certificate_id` - (Required) ID of certificate to use. Must have the domain of the service specified in the certificate's Subject Alternative Names.
* `domain` - (Required) Domain to enable TLS on. Must be assigned to an existing Fastly Service.
* `configuration_id` - (Optional) TLS Configuration to use. Defaults to the default TLS Configuration.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `created_at` - Timestamp (GMT) when TLS was enabled.

## Import

A TLS activation can be imported using its ID, e.g.

```
$ terraform import fastly_tls_activation.demo xxxxxxxx
```
