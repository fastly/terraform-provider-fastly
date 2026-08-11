---
page_title: "fastly_service_dynamic_vcl_snippet Resource - fastly"
subcategory: ""
description: |-
  Fastly dynamic VCL snippet metadata resource. Writes the versioned dynamic snippet container directly to the specified writable CDN service version.
---

# fastly_service_dynamic_vcl_snippet (Resource)

Fastly dynamic VCL snippet metadata resource. Writes the versioned dynamic
snippet container directly to the specified writable CDN service version.

This resource is part of the explicit/default first-class resource family. It
manages the versioned metadata for a dynamic VCL snippet:

- `name`
- `type`
- `priority`
- computed `snippet_id`

It does not manage dynamic snippet content. Dynamic snippet content is
versionless and is managed separately with
`fastly_service_dynamic_snippet_content`.

Use this resource when you want to manage dynamic VCL snippet metadata explicitly
against a known writable service version. For automatic service version cloning,
validation, and activation, use the nested `dynamic_snippet` block on
`fastly_service_cdn_auto`.

## Example Usage

```terraform
resource "fastly_service_cdn" "example" {
  name = "example"

  force_destroy = true
}

resource "fastly_service_dynamic_vcl_snippet" "block_scrapers" {
  service_id = fastly_service_cdn.example.id
  version    = 1

  name     = "block_scrapers"
  type     = "recv"
  priority = 100
}

resource "fastly_service_dynamic_snippet_content" "block_scrapers" {
  service_id = fastly_service_cdn.example.id
  snippet_id = fastly_service_dynamic_vcl_snippet.block_scrapers.snippet_id

  content = file("${path.module}/block_scrapers.vcl")
}
```

## Schema

### Required

- `name` (String) A name that is unique across regular and dynamic VCL snippet configuration blocks. Changing this attribute will delete and recreate the snippet.
- `service_id` (String) Fastly service ID.
- `type` (String) The location in generated VCL where the dynamic snippet should be placed. Must be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log`, or `none`.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `priority` (Number) Priority determines execution order. Lower numbers execute first. Default `100`.

### Read-Only

- `id` (String) Terraform resource identifier.
- `snippet_id` (String) The Fastly-generated dynamic snippet ID. Use this value with `fastly_service_dynamic_snippet_content` to manage versionless snippet code.

## Import

Import requires the service ID, service version, and dynamic VCL snippet name:

```shell
terraform import fastly_service_dynamic_vcl_snippet.block_scrapers SERVICE_ID/VERSION/SNIPPET_NAME
```

Example:

```shell
terraform import fastly_service_dynamic_vcl_snippet.block_scrapers SU1Z0isxPaozGVKXdv0eY/3/block_scrapers
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.

Only the dynamic snippet metadata is versioned. Dynamic snippet content is
versionless and is managed separately with
`fastly_service_dynamic_snippet_content`.

Terraform can apply metadata and content in one run when the content resource
references the computed `snippet_id`. The metadata resource still writes to the
configured writable service version, while content updates are applied directly
by snippet_id.
