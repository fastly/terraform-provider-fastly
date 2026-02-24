---
page_title: Delivery service with a versionless subdomain of a wildcard TLS subscription
subcategory: "Guides"
---

## Creating a Delivery Service using an existing Wildcard TLS Subscription

If you have an existing wildcard TLS subscription managed outside of Terraform, you can create a new Delivery service in Terraform and then link it to a subdomain of the wildcard domain.

```
terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
    }
  }
}

provider "fastly" {
  api_key = "userKey"
}

data "fastly_tls_subscription" "subscription" {
  domains               = ["*.example.com"]
  certificate_authority = "certainly"
}

resource "fastly_service_vcl" "vcl_example" {
  name          = "versionlessdomainexample"
  force_destroy = true
}

resource "fastly_domain" "domain_example" {
  fqdn       = "example.example.com"
  service_id = fastly_service_vcl.vcl_example.id
}
```