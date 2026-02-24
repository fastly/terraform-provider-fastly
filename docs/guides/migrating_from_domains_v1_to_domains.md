---
page_title: Migrating from domain_v1 objects to domain
subcategory: "Guides"
---

## Migrating 'fastly_domain_v1' and 'fastly_domain_v1_service_link' resources and data sources to 'fastly_domain' and 'fastly_domain_service_link'

As the `_v1` domain objects have become depricated, this guide will cover how you can migrate to the newer resources and data sources. 

### Step 1:
In the depircated pattern, we'll need to remove the `_v1` suffix from domain resources and data sources. 

_Before:_

```
resource "fastly_domain_v1" "example" {
    fqdn = "example.com"
    service_id = "1x2c3v4b5n6m"
    description = "This is a test domain."
}

resource "fastly_domain_v1_service_link" "example" {
    domain_id = "%s"
    service_id = "%s"
}

data "fastly_domains_v1" "example_source" {
}

```

_After:_

```
resource "fastly_domain" "example" {
    fqdn = "example.com"
    service_id = "1x2c3v4b5n6m"
    description = "This is a test domain."
}

resource "fastly_domain_service_link" "example" {
    domain_id = "%s"
    service_id = "%s"
}

data "fastly_domains" "example_source" {
}
```

### Step 2: Migrate your state file to use the new domain patterns

To ensure that your state file is aligned with the newer domain patterns, you'll need to run the following command(s) to migrate:

```
terraform state mv fastly_domain_v1.example fastly_domain.example

terraform state mv fastly_domain_v1_service_link.example fastly_domain_service_link.example

terraform state mv fastly_domain_v1.example_source fastly_domain.example_source

```

### Step 3: Confirm there are no changes / drift 

If there were no HCL changes since the last `terraform apply`, you should expect to see no diff when performing a `terraform plan` after these steps have been taken. 

Run a `terraform plan` to ensure that no drift occurs. 