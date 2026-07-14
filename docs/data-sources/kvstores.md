---
page_title: "fastly_kvstores Data Source - fastly"
subcategory: ""
description: |-
  Use this data source to retrieve a list of Fastly KV Stores.
---

# fastly_kvstores (Data Source)

Use this data source to retrieve a list of [Fastly KV Stores](https://www.fastly.com/documentation/reference/api/kv-store/).

## Example Usage

```terraform
data "fastly_kvstores" "example" {}

output "fastly_kvstores_all" {
  value = data.fastly_kvstores.example.stores
}

output "fastly_kvstores_filtered" {
  # Example: get the ID of the KV Store named "my_kv_store"
  value = one([
    for store in data.fastly_kvstores.example.stores :
    store.id if store.name == "my_kv_store"
  ])
}
```

## Schema

### Read-Only

- `id` (String) Terraform data source identifier.
- `stores` (Attributes Set) List of all KV Stores. (see [below for nested schema](#nestedatt--stores))

<a id="nestedatt--stores"></a>
### Nested Schema for `stores`

Read-Only:

- `id` (String) Identifier of the KV Store.
- `name` (String) Name of the KV Store.
