---
layout: "fastly"
page_title: "Fastly: fastly_tls_activation_ids"
sidebar_current: "docs-fastly-datasource-tls_activation_ids"
description: |-
Get the list of TLS Activation identifiers in Fastly.
---

# fastly_tls_activation_ids

Use this data source to get the list of TLS Activation identifiers in Fastly.

## Example Usage

```hcl

data "fastly_tls_activation_ids" "example" {
  certificate_id = fastly_tls_certificate.example.id
}

data "fastly_tls_activation" "example" {
  for_each = data.fastly_tls_activation_ids.example.ids
  id       = each.value
}

output "activation_domains" {
  value = [for a in data.fastly_tls_activation.example : a.domain]
}

```

## Argument Reference

* `certificate_id` - (Optional) ID of the TLS Certificate to be used with activations

## Attribute Reference

In addition to the arguments specified above, the following attributes are also supported:

* `ids` - List of IDs of the TLS Activations
