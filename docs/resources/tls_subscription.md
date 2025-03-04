---
layout: "fastly"
page_title: "Fastly: tls_subscription"
sidebar_current: "docs-fastly-resource-tls_subscription"
description: |-
Enables TLS on a domain using a managed certificate
---

# fastly_tls_subscription

Enables TLS on a domain using a certificate managed by Fastly.

DNS records need to be modified on the domain being secured, in order to respond to the ACME domain ownership challenge.

There are two options for doing this: the `managed_dns_challenges`, which is the default method; and the `managed_http_challenges`, which points production traffic to Fastly.

~> See the [Fastly documentation](https://docs.fastly.com/en/guides/serving-https-traffic-using-fastly-managed-certificates#verifying-domain-ownership) for more information on verifying domain ownership.

The examples below demonstrate usage with AWS Route53 to configure DNS, and the `fastly_tls_subscription_validation` resource to wait for validation to complete.

## Example Usage

**Basic usage:**

The following example demonstrates how to configure two subdomains (e.g. `a.example.com`, `b.example.com`).

The workflow configures a `fastly_tls_subscription` resource, then a `aws_route53_record` resource for handling the creation of the 'challenge' DNS records (e.g. `_acme-challenge.a.example.com` and `_acme-challenge.b.example.com`).

We configure the `fastly_tls_subscription_validation` resource, which blocks other resources until the challenge DNS records have been validated by Fastly.

Once the validation has been successful, the configured `fastly_tls_configuration` data source will filter the available results looking for an appropriate TLS configuration object. If that filtering process is successful, then the subsequent `aws_route53_record` resources (for configuring the subdomains) will be executed using the returned TLS configuration data.

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

**Configuring an apex and a wildcard domain:**

The following example is similar to the above but differs by demonstrating how to handle configuring an apex domain (e.g. `example.com`) and a wildcard domain (e.g. `*.example.com`) so you can support multiple subdomains to your service.

The difference in the workflow is with how to handle the Fastly API returning a single 'challenge' for both domains (e.g. `_acme-challenge.example.com`). This is done by normalising the wildcard (i.e. replacing `*.example.com` with `example.com`) and then working around the issue of the returned object having two identical keys.

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
  domains = [
    "example.com",
    "*.example.com",
  ]
}

resource "fastly_service_vcl" "example" {
  name = "example-service"

  dynamic "domain" {
    for_each = local.domains

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
    # In this example we are defining an apex (example.com) and a wildcard (*.example.com) which causes the API to return a single challenge (e.g. _acme-challenge.example.com)
    # To ensure we can match the single challenge for both domains we need to normalise the wildcard domain.
    # The 'key' is the normalised domain name (e.g. example.com)
    # The 'value' is the single 'challenge' object whose record_name matches the normalised version of the domain (e.g. record_name is _acme-challenge.example.com).
    for domain in fastly_tls_subscription.example.domains :
    replace(domain, "*.", "") => element([
      for obj in fastly_tls_subscription.example.managed_dns_challenges :
      obj if obj.record_name == "_acme-challenge.${replace(domain, "*.", "")}" # We use an `if` conditional to filter the list to a single element
    ], 0)...                                                                   # `element()` returns the first object in the list which should be the relevant 'challenge' object we need
    # The ellipsis ... avoids Terraform complaining that the resulting object will contain multiple keys that are duplicates (e.g. multiple 'example.com' keys).
    # It essentially groups the 'values' (the single challenge) under the common key (the normalised domain).
    # Then below we extract the first value (as they'll all be the same 'challenge' value).
  }

  name            = each.value[0].record_name
  type            = each.value[0].record_type
  zone_id         = aws_route53_zone.production.zone_id
  allow_overwrite = true
  records         = [each.value[0].record_value]
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

# Once validation is complete and we've retrieved the TLS configuration data, we can create multiple records...

resource "aws_route53_record" "apex" {
  name    = "example.com"
  records = [for record in data.fastly_tls_configuration.default_tls.dns_records : record.record_value if record.record_type == "A"]
  ttl     = 300
  type    = "A"
  zone_id = aws_route53_zone.production.zone_id
}

# NOTE: This subdomain matches our Fastly service because of the wildcard domain (`*.example.com`) that was added to the service.
resource "aws_route53_record" "subdomain" {
  name    = "test.example.com"
  records = [for record in data.fastly_tls_configuration.default_tls.dns_records : record.record_value if record.record_type == "CNAME"]
  ttl     = 300
  type    = "CNAME"
  zone_id = aws_route53_zone.production.zone_id
}
```

## Argument Reference

The following arguments are supported:

* `domains` - (Required) List of domains on which to enable TLS.
* `certificate_authority` - (Required) The entity that issues and certifies the TLS certificates for your subscription. Valid values are `lets-encrypt` or `globalsign`.
* `configuration_id` - (Optional) The ID of the set of TLS configuration options that apply to the enabled domains on this subscription.
* `force_update` - (Optional) Always update subscription, even when active domains are present. Defaults to false.
* `force_destroy` - (Optional) Always delete subscription, even when active domains are present. Defaults to false.

!> **Warning:** by default, the Fastly API protects you from disabling production traffic by preventing updating or deleting subscriptions with active domains. The use of `force_update` and `force_destroy` will override these protections. Take extra care using these options if you are handling production traffic.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `created_at` - Timestamp (GMT) when the subscription was created.
* `updated_at` - Timestamp (GMT) when the subscription was last updated.
* `state` - The current state of the subscription. The list of possible states are: `pending`, `processing`, `issued`, and `renewing`.
* `managed_dns_challenges` - A list of options for configuring DNS to respond to ACME DNS challenge in order to verify domain ownership. See Managed DNS Challenge below for details.
* `managed_http_challenges` - A list of options for configuring DNS to respond to ACME HTTP challenge in order to verify domain ownership. See Managed HTTP Challenges below for details.

### Managed DNS Challenge

The available attributes in the `managed_dns_challenges` block are:

* `record_name` - The name of the DNS record to add. For example `_acme-challenge.example.com`. Accessed like this, `fastly_tls_subscription.tls.managed_dns_challenges.record_name`.
* `record_type` - The type of DNS record to add, e.g. `A`, or `CNAME`.
* `record_value` - The value to which the DNS record should point, e.g. `xxxxx.fastly-validations.com`.

### Managed HTTP Challenges

The `managed_http_challenges` attribute is a set of different records that could be added depending on requirements.
For example, whether you are adding TLS to an apex domain, or a subdomain will determine which record you require.
Please note that these records will redirect production traffic to Fastly, so make sure the service is configured correctly first.
Each record in the set has the following attributes:

* `record_name` - The name of the DNS record to add. For example `example.com`. Best accessed through a `for` expression to filter the relevant record.
* `record_type` - The type of DNS record to add, e.g. `A`, or `CNAME`.
* `record_values` - A list with the value(s) to which the DNS record should point.

## Import

A subscription can be imported using its Fastly subscription ID, e.g.

```sh
$ terraform import fastly_tls_subscription.demo xxxxxxxxxxx
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `certificate_authority` (String) The entity that issues and certifies the TLS certificates for your subscription. Valid values are `lets-encrypt`, `globalsign` or `certainly`.
- `domains` (Set of String) List of domains on which to enable TLS.

### Optional

- `common_name` (String) The common name associated with the subscription generated by Fastly TLS. If you do not pass a common name on create, we will default to the first TLS domain included. If provided, the domain chosen as the common name must be included in TLS domains.
- `configuration_id` (String) The ID of the set of TLS configuration options that apply to the enabled domains on this subscription.
- `force_destroy` (Boolean) Force delete the subscription even if it has active domains. Warning: this can disable production traffic if used incorrectly. Defaults to false.
- `force_update` (Boolean) Force update the subscription even if it has active domains. Warning: this can disable production traffic if used incorrectly.

### Read-Only

- `certificate_id` (String) The certificate ID associated with the subscription.
- `created_at` (String) Timestamp (GMT) when the subscription was created.
- `id` (String) The ID of this resource.
- `managed_dns_challenge` (Map of String, Deprecated) The details required to configure DNS to respond to ACME DNS challenge in order to verify domain ownership.
- `managed_dns_challenges` (Set of Object) A list of options for configuring DNS to respond to ACME DNS challenge in order to verify domain ownership. (see [below for nested schema](#nestedatt--managed_dns_challenges))
- `managed_http_challenges` (Set of Object) A list of options for configuring DNS to respond to ACME HTTP challenge in order to verify domain ownership. Best accessed through a `for` expression to filter the relevant record. (see [below for nested schema](#nestedatt--managed_http_challenges))
- `state` (String) The current state of the subscription. The list of possible states are: `pending`, `processing`, `issued`, and `renewing`.
- `updated_at` (String) Timestamp (GMT) when the subscription was updated.

<a id="nestedatt--managed_dns_challenges"></a>
### Nested Schema for `managed_dns_challenges`

Read-Only:

- `record_name` (String)
- `record_type` (String)
- `record_value` (String)


<a id="nestedatt--managed_http_challenges"></a>
### Nested Schema for `managed_http_challenges`

Read-Only:

- `record_name` (String)
- `record_type` (String)
- `record_values` (Set of String)
