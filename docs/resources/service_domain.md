---
page_title: "fastly_service_domain Resource - fastly"
subcategory: ""
description: |-
  Fastly service domain resource. Writes directly to the specified writable service version.
---

# fastly_service_domain (Resource)

Fastly service domain resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a domain on the configured service version. It does not clone, activate,
or stage service versions.

## Example Usage

```terraform
resource "fastly_service_domain" "www" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "www.example.com"
}
```

## Schema

### Required

- `name` (String) The domain that this service responds to.
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `comment` (String) Optional comment for the domain.

### Read-Only

- `id` (String) Terraform resource identifier.

## Import

`fastly_service_domain` has a stable Framework identity of `service_id + name`.
The `version` argument is not part of the stable identity because explicit
resources can move the same logical domain from one service version to another.

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the domain from the Fastly API and
populate full state:

```shell
terraform import fastly_service_domain.www SERVICE_ID/VERSION/DOMAIN_NAME
```

Example:

```shell
terraform import fastly_service_domain.www SU1Z0isxPaozGVKXdv0eY/3/www.example.com
```

You can also use Terraform's identity-based import flow with the stable identity
fields. The resource configuration must still provide the `version` argument
because the provider needs a service version to read the domain from Fastly:

```terraform
resource "fastly_service_domain" "www" {
  service_id = "SU1Z0isxPaozGVKXdv0eY"
  version    = 3
  name       = "www.example.com"
}

import {
  to = fastly_service_domain.www

  identity = {
    service_id = "SU1Z0isxPaozGVKXdv0eY"
    name       = "www.example.com"
  }
}
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.
