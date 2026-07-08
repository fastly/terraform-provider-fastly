# Compute Resource Link Example

This example demonstrates linking an existing shared resource (such as a KV
Store or Config Store) to an explicit `fastly_service_compute` service
version via `fastly_service_resource_link`, so it's accessible from Compute
(Wasm) code at runtime.

## Usage

1. Set your Fastly API token:
   ```bash
   export FASTLY_API_TOKEN=your_token_here
   ```

2. Create the resource you want to link (e.g. a KV Store or Config Store) and
   note its ID, then set it as `linked_resource_id`, e.g. via
   `terraform.tfvars`:
   ```hcl
   linked_resource_id = "<store-id>"
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform apply
   ```

## Features

- Links a shared resource to a writable service version so it can be opened
  from Wasm code, keyed by the alias configured in `name`

## Notes

- `name` is the alias your Compute code uses to refer to the linked resource.
  It does not need to match the underlying resource's own name, and can be
  renamed in place.
- `resource_id` must reference an existing shared resource's ID. Changing it
  requires replacing the resource link, since it points the link at a
  different resource.
- `fastly_service_resource_link` writes directly to the specified service
  version; you own cloning and activating versions (see
  `fastly_service_version_clone`/`fastly_service_version_activate` actions).
