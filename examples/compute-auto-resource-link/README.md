# Compute Auto with Resource Link Example

This example demonstrates linking an existing shared resource (such as a KV
Store or Config Store) to a `fastly_service_compute_auto` service via a
`resource_link` block, so it's accessible from Compute (Wasm) code at
runtime.

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

3. Build or copy a valid Fastly Compute package to:
   ```text
   pkg/package.tar.gz
   ```

4. Initialize and apply:
   ```bash
   terraform init
   terraform apply
   ```

## Features

- Links a shared resource to the Compute service so it can be opened from
  Wasm code, keyed by the alias configured in `resource_link.name`
- `fastly_service_compute_auto` automatically clones, validates, and activates
  the service version on every change, including changes to `resource_link`
  blocks

## Notes

- `name` in the `resource_link` block is the alias your Compute code uses to
  refer to the linked resource. It does not need to match the underlying
  resource's own name.
- `resource_id` must reference an existing shared resource's ID. Changing it
  re-links to a different resource and forces a new service version.
