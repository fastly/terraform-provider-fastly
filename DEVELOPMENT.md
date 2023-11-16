## Nested Resources

Terraform doesn't have a concept of nested resources, so the Fastly Terraform provider builds an abstraction that supports this data model.

We can see an example of this nested resource design by looking at the following files (the top file being a specific resource, in this case a 'backend', for the purpose of explanation)...

- [./fastly/block_fastly_service_backend.go](./fastly/block_fastly_service_backend.go)
- [./fastly/service_crud_attribute_definition.go](./fastly/service_crud_attribute_definition.go)
- [./fastly/base_fastly_service.go](./fastly/base_fastly_service.go)

If we start at the top and look at this `backend` resource, we'll see it's actually a `TypeSet` block.

This means it's used as a 'block' inside a 'resource':

```hcl
resource fastly_service_vcl "example" {
    backend {
        name = "example-1"
    }
    backend {
        name = "example-2"
    }
}
```

The schema for the 'backend' resource is defined as a `TypeSet`, meaning each `backend` block in the above configuration is _one_ element inside of the overall set (this is important later when we cover how diffs are calculated).

We can see the constructor for the backend block (`NewServiceBackend`, defined [here](./fastly/block_fastly_service_backend.go)) actually returns something that satisfies the `ServiceAttributeDefinition` interface (code for that interface can be seen [here](./fastly/service_attribute_definition.go)).

If we look at the next file [./fastly/service_crud_attribute_definition.go](./fastly/service_crud_attribute_definition.go) we'll see that the actual concrete type returned is a `blockSetAttributeHandler` which the top-level service Resource uses to register each nested resource (this backend block being one such example), and also calls its `.Process()` method (part of the `ServiceAttributeDefinition` interface).

You can see where the service resource calls this `.Process()` method by looking at the `resourceServiceUpdate()` function in [./fastly/base_fastly_service.go](./fastly/base_fastly_service.go).

The reason `.Process()` is called only in the service's 'update' method is because the majority of nested resources can only only be created/updated when the service itself is inactive (i.e. we can make changes to the inactive service version, and then once all nested resources are created/updated, we then 'activate' the service to lock it). This means the 'update' flow for a service handles both _creation_ and _updating_ when it comes to nested resources (which of course is a little confusing at first).

The `.Process()` method uses a generic abstraction for identifying a diff for a TypeSet (which is why most 'nested resources' in the Fastly Terraform provider are built around a `TypeSet`). Once it identifies the diff between the old set and the new set, it then calls the appropriate `Delete`, `Create` or `Update` methods on the parent type (which for the sake of our example would be the backend block in [./fastly/block_fastly_service_backend.go](./fastly/block_fastly_service_backend.go)).

The important bit to realise is the diffing implementation in [./fastly/diff.go](./fastly/diff.go) works by taking the 'set' as a whole (remember this means all the 'backend' blocks defined) and converting it into a list of its elements (i.e. we'll have `[backend_block_1, backend_block_2, ...etc]`).

The diffing logic uses a `map` data structure to track each backend and it uses the `name` field of each backend to uniquely identify the backend.

So now, when the diffing logic loops over the sets (e.g. the old set and the new set) to compare differences, it will get the unique name for the current backend and use that as the 'key' in both the old and new maps.

Next, the diffing logic does two things...

1. Iterates over the difference between the new/old set and checks if the new key (i.e. the backend name) exists in the old set. If it exists in the old set, then the new set clearly contains a modified version and is added to the `modified` list. Otherwise the new set contains a newly defined backend and the backend gets added to the `added` list.
1. Iterates over the difference between the old/new (notice it's reversed) set and checks if the key (from the old set) exists in the new set. If it doesn't exist in the new set, we know this backend was deleted.

As far as updating the state file is concerned, the Read method for Terraform calls the `Read()` method for each registered nested resource. Where the `Create()`, `Update()`, `Delete()` methods of the nested resource work with a specific instance of the resource (e.g. a specific backend will be created, updated, deleted), the `Read()` method of the nested resource is responsible for calling the "List" endpoint for the resource and will get _all_ instances of a backend found via the API an flatten the data into a format that can be persisted back to the state file for the backend schema.

The `Read()` method of the nested resource is called when the service resource's `Read()` method is called (i.e. it's only called once, and not once per backend instance).
