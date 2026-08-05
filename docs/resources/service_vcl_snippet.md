---
page_title: "fastly_service_vcl_snippet Resource - fastly"
subcategory: ""
description: |-
  Fastly regular VCL snippet resource. Writes directly to the specified writable CDN service version.
---

# fastly_service_vcl_snippet (Resource)

Fastly regular VCL snippet resource. Writes directly to the specified writable CDN service version.

This resource is part of the explicit/default first-class resource family. It
manages a regular, versioned VCL snippet on the configured CDN service version.
It does not clone, activate, or stage service versions.

Use this resource when you want to manage regular VCL snippets explicitly
against a known writable service version. For automatic service version cloning,
validation, and activation, use the nested `snippet` block on
`fastly_service_cdn_auto`.

Dynamic VCL snippets are not managed by this resource. Dynamic snippets have
different lifecycle behavior because their metadata is versioned but their
content is versionless and can be updated separately.

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

resource "fastly_service_vcl_snippet" "security_headers" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "security_headers"
  type       = "deliver"
  priority   = 100
  content    = file("${path.module}/security_headers.vcl")
}
```

## Schema

### Required

- `content` (String) The VCL code that specifies exactly what the snippet does. Can be configured with a quoted string, HEREDOC, file("${path.module}/snippet.vcl"), or templatefile(...).
- `name` (String) A name that is unique across regular and dynamic VCL snippet configuration blocks. Changing this attribute will delete and recreate the snippet.
- `service_id` (String) Fastly service ID.
- `type` (String) The location in generated VCL where the snippet should be placed. Must be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log`, or `none`.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `priority` (Number) Priority determines execution order. Lower numbers execute first. Default `100`.

### Read-Only

- `id` (String) Terraform resource identifier.

## Import

Import requires the service ID, service version, and VCL snippet name:

```shell
terraform import fastly_service_vcl_snippet.security_headers SERVICE_ID/VERSION/SNIPPET_NAME
```

Example:

```shell
terraform import fastly_service_vcl_snippet.security_headers SU1Z0isxPaozGVKXdv0eY/3/security_headers
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.
