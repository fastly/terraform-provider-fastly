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

```terraform
resource "fastly_kvstore" "example" {
  name = "my_kv_store"
}
```

## Schema

### Required

- `name` (String) A unique name to identify the KV Store. Changing this attribute will delete and recreate the KV Store, discarding its current entries. Any `resource_link` referencing this KV Store must be removed first.

### Optional

- `force_destroy` (Boolean) Allow the KV Store to be deleted, even if it contains entries. Defaults to false.
- `location` (String) The regional location of the KV Store. Valid values are `US`, `EU`, `ASIA`, and `AUS`. Changing this attribute will delete and recreate the KV Store.

### Read-Only

- `id` (String) The ID of this KV Store.

## Import

Fastly KV Stores can be imported using their KV Store ID, e.g.

```shell
terraform import fastly_kvstore.example <kvstore_id>
```
