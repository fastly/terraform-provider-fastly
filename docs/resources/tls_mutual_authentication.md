---
layout: "fastly"
page_title: "Fastly: tls_mutual_authentication"
sidebar_current: "docs-fastly-resource-tls_mutual_authentication"
description: |-
Allows for client-to-server authentication using client-side X.509 authentication.
---

# fastly_tls_mutual_authentication

The Mutual TLS API allows for client-to-server authentication using client-side X.509 authentication.

The main Mutual Authentication object represents the certificate bundle and other configurations which support Mutual TLS for your domains.

Mutual TLS can be added to existing TLS activations to allow for client-to-server authentication. In order to use mutual TLS, you must already have active server-side TLS using either custom certificates or an enabled Fastly-managed subscription.

The examples below demonstrate how to use Mutual Authentication along with a TLS Subscription. Refer to the `fastly_tls_subscription` resource documentation for a deeper explanation of that code.

## Example: Single Activation

The following example sets up a TLS Subscription for `www.example.com` and then adds Mutual Authentication.

```terraform
terraform {
  required_providers {
    dnsimple = {
      source  = "dnsimple/dnsimple"
      version = "1.5.0"
    }
    fastly = {
      source  = "fastly/fastly"
      version = "5.7.2"
    }
  }
}

variable "dnsimple_token" {
  type = string
}

variable "dnsimple_account" {
  type = string
}

provider "dnsimple" {
  account = var.dnsimple_account
  token   = var.dnsimple_token
}

variable "zone" {
  type    = string
  default = "example.com"
}

resource "fastly_service_vcl" "example" {
  name = "example"
  domain {
    name = "www.${var.zone}"
  }
  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }
  force_destroy = true
}

resource "fastly_tls_subscription" "www" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name if domain.name == "www.${var.zone}"]
  certificate_authority = "certainly"
}

resource "dnsimple_zone_record" "www_acme_challenge" {
  name      = "_acme-challenge.www"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for obj in fastly_tls_subscription.www.managed_dns_challenges : obj.record_value if obj.record_name == "_acme-challenge.www.${var.zone}"])
  zone_name = var.zone
}

resource "fastly_tls_subscription_validation" "www" {
  subscription_id = fastly_tls_subscription.www.id
  depends_on      = [dnsimple_zone_record.www_acme_challenge]
}

data "fastly_tls_configuration" "default" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.www]
}

resource "dnsimple_zone_record" "www" {
  name      = "www"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

data "fastly_tls_activation" "www" {
  domain     = "www.example.com"
  depends_on = [dnsimple_zone_record.www]
}

resource "fastly_tls_mutual_authentication" "www" {
  activation_ids = [data.fastly_tls_activation.www.id]
  cert_bundle    = "-----BEGIN CERTIFICATE-----\n<REDACTED>\n-----END CERTIFICATE-----"
  enforced       = true
}
```

## Example: Multiple Activations

The following example sets up a TLS Subscription for `foo.example.com` and `bar.example.com` and then adds Mutual Authentication to each TLS Activation.

```terraform
terraform {
  required_providers {
    dnsimple = {
      source  = "dnsimple/dnsimple"
      version = "1.5.0"
    }
    fastly = {
      source  = "fastly/fastly"
      version = "5.7.2"
    }
  }
}

variable "dnsimple_token" {
  type = string
}

variable "dnsimple_account" {
  type = string
}

provider "dnsimple" {
  account = var.dnsimple_account
  token   = var.dnsimple_token
}

variable "zone" {
  type    = string
  default = "example.com"
}

resource "fastly_service_vcl" "example" {
  name = "example"
  domain {
    name = "foo.${var.zone}"
  }
  domain {
    name = "bar.${var.zone}"
  }
  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }
  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name]
  certificate_authority = "certainly"
}

resource "dnsimple_zone_record" "example_acme_challenge" {
  for_each = {
    for domain in fastly_tls_subscription.example.domains : domain => one([
      for obj in fastly_tls_subscription.example.managed_dns_challenges : obj if obj.record_name == "_acme-challenge.${domain}"
    ])
  }
  name      = each.value.record_name
  ttl       = "60"
  type      = each.value.record_type
  value     = each.value.record_value
  zone_name = var.zone
}

resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [dnsimple_zone_record.example_acme_challenge]
}

data "fastly_tls_configuration" "default" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.example]
}

resource "dnsimple_zone_record" "foo" {
  name      = "foo"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

resource "dnsimple_zone_record" "bar" {
  name      = "bar"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

data "fastly_tls_activation_ids" "example" {
  certificate_id = fastly_tls_subscription.example.certificate_id
}

resource "fastly_tls_mutual_authentication" "example" {
  activation_ids = data.fastly_tls_activation_ids.example.ids
  cert_bundle    = "-----BEGIN CERTIFICATE-----\n<REDACTED>\n-----END CERTIFICATE-----"
  enforced       = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cert_bundle` (String) One or more certificates. Enter each individual certificate blob on a new line. Must be PEM-formatted.

### Optional

- `activation_ids` (Set of String) List of TLS Activation IDs
- `enforced` (Boolean) Determines whether Mutual TLS will fail closed (enforced) or fail open. A true value will require a successful Mutual TLS handshake for the connection to continue and will fail closed if unsuccessful. A false value will fail open and allow the connection to proceed (if this attribute is not set we default to `false`).
- `include` (String) A comma-separated list used by the Terraform provider during a state refresh to return more data related to your mutual authentication from the Fastly API (permitted values: `tls_activations`).
- `name` (String) A custom name for your mutual authentication. If name is not supplied we will auto-generate one.

### Read-Only

- `created_at` (String) Date and time in ISO 8601 format.
- `id` (String) The ID of this resource.
- `tls_activations` (List of String) List of alphanumeric strings identifying TLS activations.
- `updated_at` (String) Date and time in ISO 8601 format.
