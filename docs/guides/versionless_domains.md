---
page_title: versionless_domains
subcategory: "Guides"
---

## Versionless Domains

There are several different ways your Fastly services can use versionless domains

## Delivery Service

If you have an existing wildcard TLS subscription managed outside of Terraform, you can create a new subdomain in Terraform and then link it to a Delivery service

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
  domains               = ["*.fastlyversionlessdomain.com"]
  certificate_authority = "certainly"
}

resource "fastly_domain" "domain_demo" {
  fqdn = "example.fastlyversionlessdomain.com"
}

resource "fastly_service_vcl" "vcl_demo" {
  name          = "demofastly"
  force_destroy = true
}

resource "fastly_domain_service_link" "link" {
  domain_id  = fastly_domain.domain_demo.id
  service_id = fastly_service_vcl.vcl_demo.id
}
```