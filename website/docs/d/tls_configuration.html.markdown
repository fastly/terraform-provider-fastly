---
layout: "fastly"
page_title: "Fastly: fastly_tls_configuration"
sidebar_current: "docs-fastly-datasource-tls_configuration"
description: |-
Get information on Fastly TLS configuration.
---

# fastly_tls_configuration

Use this data source to get the ID of a TLS configuration for use with other resources.

## Example Usage

```hcl
data "fastly_tls_configuration" "example" {
  default = true
}

resource "fastly_tls_activation" "example" {
  configuration_id = data.fastly_tls_configuration.example.id
  // ...
}
```

## Argument Reference

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination with any of the others.
* `id` - (Optional) ID of the TLS configuration obtained from the Fastly API or another data source. Conflicts with all the other filters.
* `name` - (Optional) Custom name of the TLS configuration 
* `tls_protocols` - (Optional) TLS protocols available on the TLS configuration
* `http_protocols` - (Optional) HTTP protocols available on the TLS configuration
* `tls_service` - (Optional) Whether the configuration should support the `PLATFORM` or `CUSTOM` TLS service
* `default` - (Optional) Signifies whether Fastly will use this configuration as a default when creating a new TLS activation

## Attribute Reference

* `created_at` - Time-stamp (GMT) when the configuration was created
* `updated_at` - Time-stamp (GMT) when the configuration was last updated
* `dns_records` - The DNS records to use for the configuration. See DNS Records below for details.

### DNS Records

* `record_type` - Type of DNS record to set, e.g. A, AAAA, or CNAME.
* `record_value` - The IP address or hostname of the DNS record.
* `region` - The regions that will be used to route traffic. Select DNS Records with a `global` region to route traffic to the most performant point of presence (POP) worldwide (global pricing will apply). Select DNS records with a `us-eu` region to exclusively land traffic on North American and European POPs.
