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

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.55.0"
    }
    fastly = {
      source  = "fastly/fastly"
      version = "3.1.0"
    }
  }
}

# NOTE: Creating a hosted zone will automatically create SOA/NS records.
resource "aws_route53_zone" "production" {
  name = "example.com"
}

resource "aws_route53domains_registered_domain" "example" {
  domain_name = "example.com"

  dynamic "name_server" {
    for_each = aws_route53_zone.production.name_servers

    content {
      name = name_server.value
    }
  }
}

locals {
  subdomains = [
    "a.example.com",
    "b.example.com",
  ]
}

resource "fastly_service_vcl" "example" {
  name = "example-service"

  dynamic "domain" {
    for_each = local.subdomains

    content {
      name = domain.value
    }
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

resource "aws_route53_record" "domain_validation" {
  depends_on = [fastly_tls_subscription.example]

  for_each = {
    # The following `for` expression (due to the outer {}) will produce an object with key/value pairs.
    # The 'key' is the domain name we've configured (e.g. a.example.com, b.example.com)
    # The 'value' is a specific 'challenge' object whose record_name matches the domain (e.g. record_name is _acme-challenge.a.example.com).
    for domain in fastly_tls_subscription.example.domains :
    domain => element([
      for obj in fastly_tls_subscription.example.managed_dns_challenges :
      obj if obj.record_name == "_acme-challenge.${domain}" # We use an `if` conditional to filter the list to a single element
    ], 0)                                                   # `element()` returns the first object in the list which should be the relevant 'challenge' object we need
  }

  name            = each.value.record_name
  type            = each.value.record_type
  zone_id         = aws_route53_zone.production.zone_id
  allow_overwrite = true
  records         = [each.value.record_value]
  ttl             = 60
}

# This is a resource that other resources can depend on if they require the certificate to be issued.
# NOTE: Internally the resource keeps retrying `GetTLSSubscription` until no error is returned (or the configured timeout is reached).
resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [aws_route53_record.domain_validation]
}

# This data source lists all available configuration objects.
# It uses a `default` attribute to narrow down the list to just one configuration object.
# If the filtered list has a length that is not exactly one element, you'll see an error returned.
# The single TLS configuration is then returned and can be referenced by other resources (see aws_route53_record below).
#
# IMPORTANT: Not all customers will have a 'default' configuration.
# If you have issues filtering with `default = true`, then you may need another attribute.
# Refer to the fastly_tls_configuration documentation for available attributes:
# https://registry.terraform.io/providers/fastly/fastly/latest/docs/data-sources/tls_configuration#optional
data "fastly_tls_configuration" "default_tls" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.example]
}

# Once validation is complete and we've retrieved the TLS configuration data, we can create multiple subdomain records.
resource "aws_route53_record" "subdomain" {
  for_each = toset(local.subdomains) # Because `subdomains` is ultimately a list, the `each` variable produced will contain only a `value` property which will be the subdomain.

  name    = each.value # e.g. a.example.com, b.example.com
  records = [for record in data.fastly_tls_configuration.default_tls.dns_records : record.record_value if record.record_type == "CNAME"]
  ttl     = 300
  type    = "CNAME"
  zone_id = aws_route53_zone.production.zone_id
}
```

Managed certificates for multiple domains, in a single apply:

The `certificate_id` attribute is only populated once the certificate has been issued, so resources referencing it are guaranteed to run after issuance — unlike `fastly_tls_subscription.<name>.certificate_id`, which is empty on first apply (certificates are issued asynchronously after domain validation) and causes API 400 errors when consumed in the same apply.

~> **Note:** Fastly automatically activates TLS on a subscription's domains once the certificate is issued — set `configuration_id` on the `fastly_tls_subscription` itself and do **not** create a `fastly_tls_activation` for those domains (it fails with `400 domain_id has already been taken`). Use the `fastly_tls_activation` data source to read the automatically-created activation.

```terraform
# Certificates map: key = domain (common_name), value = subscription config
variable "certificates" {
  type = map(object({
    authority     = optional(string, "lets-encrypt")
    common_name   = optional(string, "")
    domains       = set(string)
    force_destroy = optional(bool, true)
  }))
}

data "fastly_tls_configuration" "default_tls" {
  default = true
}

resource "fastly_tls_subscription" "example" {
  for_each = var.certificates

  certificate_authority = each.value.authority
  common_name           = each.value.common_name
  domains               = each.value.domains
  force_destroy         = each.value.force_destroy

  # IMPORTANT: with a managed subscription, set the TLS configuration HERE.
  # Fastly automatically activates TLS on the subscription's domains once the
  # certificate is issued. Do NOT create a fastly_tls_activation for these
  # domains — it will fail with `400 domain_id has already been taken`.
  configuration_id = data.fastly_tls_configuration.default_tls.id
}

# The domain validation challenge records MUST be created before validation
# can succeed. This example uses DNS-based validation: replace with your DNS
# provider's record resource, fed from
# fastly_tls_subscription.example[each.key].managed_dns_challenges
# (see the fastly_tls_subscription documentation for a Route53 example).
resource "aws_route53_record" "domain_validation" {
  for_each = fastly_tls_subscription.example

  name            = each.value.managed_dns_challenges[0].record_name
  type            = each.value.managed_dns_challenges[0].record_type
  records         = [each.value.managed_dns_challenges[0].record_value]
  zone_id         = "REPLACE_WITH_YOUR_ZONE_ID"
  allow_overwrite = true
  ttl             = 60
}

# Blocks until the certificate has been issued. Its certificate_id attribute
# is only known after issuance, so downstream resources referencing it are
# guaranteed to run with a valid certificate — all in a single apply.
resource "fastly_tls_subscription_validation" "example" {
  for_each = var.certificates

  subscription_id = fastly_tls_subscription.example[each.key].id
  depends_on      = [aws_route53_record.domain_validation]
}

# The activation was created automatically by Fastly when the certificate was
# issued; this data source reads it (e.g. to consume dns_records/IDs).
data "fastly_tls_activation" "example" {
  for_each = var.certificates

  domain     = each.value.common_name
  depends_on = [fastly_tls_subscription_validation.example]
}

output "certificate_ids" {
  value = { for k, v in fastly_tls_subscription_validation.example : k => v.certificate_id }
}
```

## Timeouts

`fastly_tls_subscription_validation` supports the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `45m`) How long to wait for the subscription to be validated.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `subscription_id` (String) The ID of the TLS Subscription that should be validated.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `certificate_id` (String) The ID of the certificate issued for the validated subscription. Only populated once the subscription reaches the `issued` state. Reference this from `fastly_tls_activation.certificate_id` to guarantee the activation is created after the certificate exists, within a single apply.
- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
