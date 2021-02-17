---
layout: "fastly"
page_title: "Fastly: tls_subscription_validation"
sidebar_current: "docs-fastly-resource-tls_subscription_validation"
description: |-
Represents a successful validation of a Fastly TLS Subscription
---

# fastly_tls_subscription_validation

This resource represents a successful validation of a Fastly TLS Subscription in concert with other resources.

Most commonly, this resource is used together with a resource for a DNS record and `fastly_tls_subscription` to request a DNS validated certificate, deploy the required validation records and wait for validation to complete.

~> **Warning:** This resource implements a part of the validation workflow. It does not represent a real-world entity in Fastly, therefore changing or deleting this resource on its own has no immediate effect.

## Example Usage

DNS Validation with AWS Route53:

```hcl
locals {
  domain_name = "example.com"
  challenge_fields = [for c in fastly_tls_subscription.example.tls_authorization_challenges : {
    name    = c.record_name
    type    = c.record_type
    records = c.record_values
  } if c["challenge_type"] == "managed-dns"][0]
}

data "aws_route53_zone" "test" {
  name         = local.domain_name
  private_zone = false
}

resource "aws_route53_record" "domain_validation" {
  name            = local.challenge_fields.name
  type            = local.challenge_fields.type
  zone_id         = data.aws_route53_zone.test.id
  allow_overwrite = true
  records         = local.challenge_fields.records
  ttl             = 60
}

resource "fastly_service_v1" "example" {
  name = "example-service"

  domain {
    name = local.domain_name
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_v1.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [aws_route53_record.domain_validation]
}
```

## Argument Reference

The following arguments are supported:

* `subscription_id` - (Required) The ID of the TLS Subscription that should be validated.

## Attributes Reference

No other attributes are available for this resource.

## Timeouts

`fastly_tls_subscription_validation` supports the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `45m`) How long to wait for the subscription to be validated.