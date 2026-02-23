---
page_title: Delivery service with a versionless domain and TLS subscription using Certainly
subcategory: "Guides"
---

## Creating a Delivery Service with a Certainly TLS Subscription using a versionless domain

The following guide exemplifies how to use versionless domains with a Certainly subscription to link to a new delivery service. 

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

resource "fastly_service_vcl" "example_delivery_service" {
  name = "tls-service-2"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_domain" "domain_example" {
  description = "My Fastly example domain"
  fqdn        = "fastly.example.com"
  service_id  = fastly_service_vcl.example_delivery_service.id
}


resource "fastly_tls_subscription" "subscription" {
  domains               = [fastly_domain.domain_example.fqdn]
  certificate_authority = "certainly"
  
  // The 'depends_on' attribute ensures that the 
  // TLS subscription is created after the domain.
  depends_on = [fastly_domain.domain_example]
}

```


## Creating a Delivery Service with a Certainly TLS Subscription for bulk domains

The following guide provides an example of how to bulk manage domains through a Certainly subscription. 

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

locals {
  subdomains = [
    "dev.fastly.example.com",
    "nonprod.fastly.example.com",
    "prod.fastly.example.com",
    "staging.fastly.example.com",
    "staging2.fastly.example.com",
    "staging3.fastly.example.com"
  ]
}

resource "fastly_service_vcl" "bulk_example_delivery_service" {
  name = "tls-service-2"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_domain" "bulk_domains" {
    for_each = toset(local.subdomains)

    description = "Bulk managed Fastly domain"
    fqdn        = each.value
    service_id  = fastly_service_vcl.bulk_example_delivery_service.id
}

resource "fastly_tls_subscription" "subscription" {
    domains               = local.subdomains
    certificate_authority = "certainly"

    // The 'depends_on' attribute ensures that the 
    // TLS subscription is created after the domain.
    depends_on = [fastly_domain.bulk_domains]
}

```

