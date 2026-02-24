---
page_title: Migrate a Delivery service with a classic domain to a versionless domain
subcategory: "Guides"
---

## Migrate a Delivery service with a classic domain to a versionless domain

Before migrating, your HCL will look something like this, with a domain block inside of your service.

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

resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  domain {
    name    = "demo.example.com"
    comment = "demo"
  }
  force_destroy = true
}
```

Once you [use the control panel to migrate this domain to a versionless domain](https://www.fastly.com/documentation/guides/getting-started/domains/working-with-domains/migrating-classic-domains/), you will need to update your HCL to look something like this.
[Here are the versionless domain docs.](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/domain)

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

resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  force_destroy = true
}

resource "fastly_domain" "domain_example" {
  fqdn       = "demo.example.com"
  service_id = fastly_service_vcl.vcl_example.id
}
```

Before making other changes you will need to import domain using terraform.

```
terraform import fastly_domain.domain_example <domain_id>
```

After that, running `terraform plan` should result in no changes