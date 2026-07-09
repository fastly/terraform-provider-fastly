---
page_title: "fastly_service_resource_link Resource - fastly"
subcategory: ""
description: |-
  Links a shared resource (such as a KV Store or Config Store) to a Fastly service version, making it accessible from Compute code. Writes directly to the specified writable service version.
---

# fastly_service_resource_link (Resource)

Links a shared resource (such as a KV Store or Config Store) to a Fastly service version, making it accessible from Compute code. Writes directly to the specified writable service version.

## Example Usage

```terraform
resource "fastly_service_compute" "app" {
  name    = "example-compute-service"
  comment = "Managed by Terraform"
}

# Makes the linked resource available to Wasm code, e.g. as a KV Store or
# Config Store lookup keyed by "store".
resource "fastly_service_resource_link" "store" {
  service_id  = fastly_service_compute.app.id
  version     = 1
  name        = "store"
  resource_id = var.linked_resource_id
}
```

## Schema

### Required

- `name` (String) The name the service will use to open the linked resource from Compute code (e.g. a KV Store or Config Store SDK lookup). This is an alias and does not need to match the name of the underlying resource.
- `resource_id` (String) The ID of the shared resource to link (e.g. the ID of a KV Store or Config Store).
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Read-Only

- `id` (String) Terraform resource identifier.
- `link_id` (String) An alphanumeric string identifying this resource link.

## Import

The import ID format is `service_id/version/name`.

```shell
terraform import fastly_service_resource_link.store SERVICE_ID/VERSION/NAME
```

Example:

```shell
terraform import fastly_service_resource_link.store SU1Z0isxPaozGVKXdv0eY/3/store
```
