---
page_title: Migrate a Delivery service with a classic domain to a versionless domain
subcategory: "Guides"
---

## Migrate a Delivery service with a classic domain to a versionless domain

Before migrating, your HCL will look something like this, with a domain block inside of your service and a TLS subscription.

```
resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }
  domain {
    name    = "demo.example.com"
    comment = "demo"
  }
}

resource "fastly_tls_subscription" "migrate_example" {
  domains = ["demo.example.com"]
  certificate_authority = "lets-encrypt"
}
```

Before making other changes you will need to import your domain(s) using Terraform. We'll need to run the following import command for each domain:
> `terraform import fastly_domain.domain_example <domain_id>` 

The Domain ID can be obtained by using [the Fastly CLI](https://www.fastly.com/documentation/reference/tools/cli/) by running `fastly domain list --fqdn=foo.example.com` and then using the Domain ID from the record.

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

You'll then need to add a `fastly_domain` resource, associating your service, as seen below.   
  _You can find the documentation on versionless domain patterns here: https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/domain._ 
> **Do not apply your changes before running the import in the next step, as that will result in an error.**

```
resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }
  domain {
    name    = "demo.example.com"
    comment = "demo"
  }
}

resource "fastly_tls_subscription" "migrate_example" {
  domains = ["demo.example.com"]
  certificate_authority = "lets-encrypt"
}

resource "fastly_domain" "domain_example" {
  fqdn       = "demo.example.com"
  service_id = fastly_service_vcl.vcl_example.id
}
```

Once you have added the necessary `fastly_domain` blocks , you can proceed with a `terraform plan` / `terraform apply`. 

You should excpect to see something like this from your `terraform plan` / `terraform apply` commands:
```
 # fastly_domain.example will be updated in-place
  ~ resource "fastly_domain" "example" {
        id          = "3a2b3c4d5e"
      + service_id  = "1a2b4c5d6e"
        # (3 unchanged attributes hidden)
    }
```

You now have associated the versionless domain with your service, but we'll still need to remove the classic domain from your HCL. The final step is to remove the `domain` attribute from your `fastly_service_vcl` block, as seen below.

```
resource "fastly_service_vcl" "vcl_example" {
  name = "versionlessdomainexample"
  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }
}

resource "fastly_tls_subscription" "migrate_example" {
  domains = ["demo.example.com"]
  certificate_authority = "lets-encrypt"
}

resource "fastly_domain" "domain_example" {
  fqdn       = "demo.example.com"
  service_id = fastly_service_vcl.vcl_example.id
}
```

You can now perform the `terraform plan` / `terraform apply` commands. You should expect to see a similar output from Terraform:

```
  # fastly_service_vcl.migrate_example will be updated in-place
  ~ resource "fastly_service_vcl" "migrate_example" {
      ~ active_version     = 1 -> (known after apply)
      ~ cloned_version     = 1 -> (known after apply)
        id                 = "1a2b4c5d6e"
        name               = "versionlessdomainexample"
        # (11 unchanged attributes hidden)

      - domain {
          - name    = "demo.example.com" -> null
            # (1 unchanged attribute hidden)
        }

        # (1 unchanged block hidden)
    }
```
