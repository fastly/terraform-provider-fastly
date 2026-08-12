---
page_title: "fastly_vcl_snippets Data Source - fastly"
subcategory: ""
description: |-
  Use this data source to retrieve VCL snippets for a Fastly service version.
---

# fastly_vcl_snippets (Data Source)

Use this data source to retrieve regular and dynamic VCL snippet metadata for a Fastly service version.

This data source reads snippets through the versioned snippet list endpoint. It is useful with both VCL snippet resource families added in:

- regular VCL snippets: [#1381](https://github.com/fastly/terraform-provider-fastly/pull/1381)
- dynamic VCL snippets: [#1397](https://github.com/fastly/terraform-provider-fastly/pull/1397)

Dynamic snippet content is versionless and is managed separately with `fastly_service_dynamic_snippet_content`.

## Example Usage

```terraform
data "fastly_vcl_snippets" "example" {
  service_id      = fastly_service_cdn_auto.example.id
  service_version = fastly_service_cdn_auto.example.active_version
}

output "recv_snippet_names" {
  value = [
    for snippet in data.fastly_vcl_snippets.example.vcl_snippets :
    snippet.name if snippet.type == "recv"
  ]
}
```

## Schema

### Required

- `service_id` (String) Fastly service ID.
- `service_version` (Number) Fastly service version to read snippets from.

### Read-Only

- `id` (String) Terraform data source identifier.
- `vcl_snippets` (Attributes Set) List of all VCL snippets for the configured service version. (see [below for nested schema](#nestedatt--vcl_snippets))

<a id="nestedatt--vcl_snippets"></a>
### Nested Schema for `vcl_snippets`

Read-Only:

- `content` (String) The VCL code that specifies exactly what the snippet does.
- `dynamic` (Boolean) Whether this is a dynamic VCL snippet.
- `id` (String) Fastly-generated VCL snippet identifier.
- `name` (String) The name for the snippet.
- `priority` (Number) Priority determines execution order. Lower numbers execute first.
- `type` (String) The location in generated VCL where the snippet should be placed.
