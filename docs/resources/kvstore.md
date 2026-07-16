---
page_title: "fastly_kvstore Resource - fastly"
subcategory: ""
description: |-
  Provides a KV Store, a low-latency, high-throughput key-value data store that is accessible to Compute services during request processing.
---

# fastly_kvstore (Resource)

Provides a KV Store, a low-latency, high-throughput key-value data store that is accessible to Compute services during request processing.

This resource is versionless: it is not tied to a service version and is managed independently of any `fastly_service_compute` resource.

## Example Usage

Basic usage:

```terraform
resource "fastly_kvstore" "example" {
  name = "my_kv_store"
}
```

Linking the store to a Compute service so it's readable from Wasm code, using
the `resource_link` block nested in `fastly_service_compute_auto`:

```terraform
resource "fastly_kvstore" "example" {
  name     = "my_kv_store"
  location = "US"
}

resource "fastly_service_compute_auto" "example" {
  name = "my_compute_service"

  domain {
    name = "demo.example.com"
  }

  package {
    filename         = "package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  resource_link {
    name        = "my_resource_link"
    resource_id = fastly_kvstore.example.id
  }
}

data "fastly_package_hash" "example" {
  filename = "package.tar.gz"
}
```

For explicit (versioned) Compute services, use the standalone
`fastly_service_resource_link` resource instead of a nested block:

```terraform
resource "fastly_kvstore" "example" {
  name = "my_kv_store"
}

resource "fastly_service_compute" "example" {
  name    = "my_compute_service"
  comment = "Managed by Terraform"
}

resource "fastly_service_resource_link" "example" {
  service_id  = fastly_service_compute.example.id
  version     = 1
  name        = "my_resource_link"
  resource_id = fastly_kvstore.example.id
}
```

-> **Note:** A KV Store cannot be deleted while a `resource_link` still
references it. Since Terraform cannot guarantee ordering between a
`fastly_kvstore` and the service resource linking to it, removing both in the
same `terraform apply` may require running `apply` twice: once to remove the
`resource_link`, and again to delete the KV Store.

## Schema

### Required

- `name` (String) A unique name to identify the KV Store. Changing this attribute will delete and recreate the KV Store, discarding its current entries. Any `resource_link` referencing this KV Store must be removed first.

### Optional

- `force_destroy` (Boolean) Allow the KV Store to be deleted, even if it contains entries. Defaults to false.
- `location` (String) The regional location of the KV Store. Valid values are `US`, `EU`, `ASIA`, and `AUS`. Changing this attribute will delete and recreate the KV Store. The Fastly API does not return the configured location, so it cannot be verified on `terraform import`.

### Read-Only

- `id` (String) The ID of this KV Store.

## Import

Fastly KV Stores can be imported using their KV Store ID, e.g.

```shell
terraform import fastly_kvstore.example <kvstore_id>
```
