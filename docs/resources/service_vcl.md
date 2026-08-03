---
page_title: "fastly_service_vcl Resource - fastly"
subcategory: ""
description: |-
  Fastly custom VCL file resource. Writes directly to the specified writable CDN service version.
---

# fastly_service_vcl (Resource)

Fastly custom VCL file resource. Writes directly to the specified writable CDN service version.

This resource is part of the explicit/default first-class resource family. It
manages a custom VCL file on the configured CDN service version. It does not
clone, activate, or stage service versions.

Use this resource when you want to manage custom VCL files explicitly against a
known writable service version. For automatic service version cloning,
validation, and activation, use the nested `vcl` block on
`fastly_service_cdn_auto`.

## Example Usage

```terraform
resource "fastly_service_cdn" "example" {
  name = "example"

  domain {
    name = "www.example.com"
  }

  backend {
    name    = "origin"
    address = "origin.example.com"
    port    = 443
    use_ssl = true
  }

  force_destroy = true
}

resource "fastly_service_vcl" "main" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "main"
  main       = true
  content    = file("${path.module}/main.vcl")
}
```

## Schema

### Required

- `content` (String) The custom VCL source code to upload. Can configured with file("${path.module}/main.vcl") or templatefile(...).
- `name` (String) A unique name for this custom VCL file. Included VCL files must be referenced by this exact name from the main VCL file.
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `main` (Boolean) Whether this custom VCL file is the main configuration. Exactly one configured custom VCL file must be marked as main.

### Read-Only

- `id` (String) Terraform resource identifier.

## Import

Import requires the service ID, service version, and custom VCL file name:

```shell
terraform import fastly_service_vcl.main SERVICE_ID/VERSION/VCL_NAME
```

Example:

```shell
terraform import fastly_service_vcl.main SU1Z0isxPaozGVKXdv0eY/3/main
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.
