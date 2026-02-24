---
page_title: Migrate a Delivery service with a classic domain to a versionless domain
subcategory: "Guides"
---

## Migrate a Delivery service with a classic domain to a versionless domain

Before migrating, your HCL will look something like this, with a domain block inside of your service.

```
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
You can find the documentation on versionless domain patterns here: https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/domain. Do not apply your changes before running the import step below, as that will result in an error.

```
resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  force_destroy = true
}

resource "fastly_domain" "domain_example" {
  fqdn       = "demo.example.com"
}
```

Before making other changes you will need to import domain using terraform. The Domain ID can using the Fastly CLI by running `fastly domain list --fqdn=foo.example.com` and using the Domain ID fromt the record.

```
terraform import fastly_domain.domain_example YOUR_DOMAIN_ID
```

Terraform should produce a message similar to the following.

```
fastly_domain.example: Importing from ID "YOUR_DOMAIN_ID"...
fastly_domain.example: Import prepared!
  Prepared fastly_domain for import
fastly_domain.example: Refreshing state... [id=YOUR_DOMAIN_ID]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

After importing, running `terraform plan` should result in no changes.

Once the import has run successfully, you can assign the domain to a service using the `service_id` field of the domain and then plan and apply.

```
resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  force_destroy = true
}

resource "fastly_domain" "domain_example" {
  fqdn       = "demo.example.com"
  service_id = fastly_service_vcl.vcl_example.id
}
```