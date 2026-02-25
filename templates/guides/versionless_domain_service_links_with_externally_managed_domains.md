---
page_title: Linking versionless domains to a service with externally managed domains
subcategory: "Guides"
---

## Linking versionless domains that are externally managed to a delivery service

The following guide goes over how you would link externally-managed versionless domains to a given service. The domains may be managed using Terraform in a separate configuration, or managed using other tools or methods.

_Note: These domains must already exist in your Fastly account / configuration prior in order for this pattern to be successful_ 


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
  # First group of domains
  exampleProperties = [
    "dev.fastly.example.com",
    "nonprod.fastly.example.com",
    "prod.fastly.example.com"
  ]

  # Second group of domains
  sampleProperties = [
    "dev.fastly.sample.com",
    "nonprod.fastly.sample.com",
    "prod.fastly.sample.com"
  ]

  # Create domain ID maps for each set
  exampleDomainIds = {
    for domain in data.fastly_domains.all.domains :
    domain.fqdn => domain.id
    if contains(local.exampleProperties, domain.fqdn)
  }

  sampleDomainIds = {
    for domain in data.fastly_domains.all.domains :
    domain.fqdn => domain.id
    if contains(local.sampleProperties, domain.fqdn)
  }
}

# Fetch all domains
data "fastly_domains" "all" {}


resource "fastly_service_vcl" "linking_service1" {
  name = "example service"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }
}

resource "fastly_service_vcl" "linking_service2" {
  name = "sample service"

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }
}

# Link example domains to linking_service1
resource "fastly_domain_service_link" "example_links" {
  for_each = local.exampleDomainIds

  domain_id  = each.value
  service_id = fastly_service_vcl.linking_service1.id
}

# Link sample domains to linking_service2
resource "fastly_domain_service_link" "sample_links" {
  for_each = local.sampleDomainIds

  domain_id  = each.value
  service_id = fastly_service_vcl.linking_service2.id
}

  ```
