# Orchestration Example (`fastly_service_cdn_auto`)

This example demonstrates the automatic compatibility workflow in the dual-model
Fastly Terraform provider. It uses `fastly_service_cdn_auto` with nested
`domain` and `backend` blocks.

The example provisions and manages:

- two Fastly CDN services
- one shared backend reused across both services
- one service-specific backend on only service 1
- one domain per service
- one ACL per service (service 1 has an IP allowlist, service 2 has a temporary blocklist)
- one gzip configuration on service 1
- Image Optimizer enabled on service 1 (default settings added in a follow-up apply - see below)

## What this example is for

This example tests current-provider-style behavior with the automatic
compatibility resource family:

- one aggregate service resource owns the nested service configuration
- nested domain/backend changes are reconciled at the service level
- one working version is selected per changed service update
- the provider validates and activates automatically

## Files

```text
examples/orchestration-cdn-auto/
  main.tf
  variables.tf
  terraform.tfvars
  outputs.tf
  README.md
```

## One-time bootstrap

For a brand new environment:

```bash
cd examples/orchestration-cdn-auto
terraform init
terraform apply
```

Expected bootstrap behavior:

1. Terraform creates both `fastly_service_cdn_auto` services.
2. Fastly creates editable version `1` for each new service.
3. The provider reconciles the nested `domain`, `backend`, `acl`, and `gzip`
   blocks to version `1`.
4. The provider validates and activates version `1`.
5. `fastly_service_product_image_optimizer.service_1` enables the Image
   Optimizer product on `service_1`.

Note: `image_optimizer_default_settings` can't be added to `service_1` in
this same bootstrap apply. That block is reconciled as part of
`fastly_service_cdn_auto.service_1`'s own create step, which runs before
`fastly_service_product_image_optimizer.service_1` - so Image Optimizer isn't
enabled yet at the point the settings block would need it. See "Configure
Image Optimizer default settings" below for the required follow-up apply.

## Day-to-day workflow

Make configuration changes in `main.tf`, then run:

```bash
terraform fmt -recursive
terraform validate
terraform plan
terraform apply
```

Expected update behavior for a changed service:

1. the provider selects one working version for that service
2. it reconciles all nested versioned resources to that working version
3. it validates the working version
4. it activates the working version automatically

If only `service_1` changes, only `service_1` should get a new managed and active
version. `service_2` should remain unchanged.

## Suggested manual tests

### Change only service 1

For example, change the port of the unique backend on service 1 from `80` to `443`.

Then run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_active_version` move to the new version
- `service_2_*` outputs remain unchanged

### Change both services

For example:

- change the shared backend port
- or change both domain names

Then run:

```bash
terraform apply
```

You should see both services get new versions and activation.

### Delete a nested backend

For example, remove the unique backend from `service_1` and run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_active_version` move to the new version
- the removed backend disappear from the service configuration
- `service_2_*` outputs remain unchanged

A follow-up `terraform plan` should show no changes.

### Add or modify an ACL

For example, add a second ACL to `service_1` or change the `force_destroy` attribute on `service_2`'s ACL:

```hcl
  acl {
    name = "secondary_allowlist"
  }
```

Then run:

```bash
terraform apply
```

You should see:

- the affected service get a new managed version
- the new or modified ACL appear in the service configuration
- `acl_id` outputs reflect the new ACL IDs

### Add or modify a gzip configuration

For example, remove `application/javascript` from service 1's gzip `content_types`:

```hcl
  gzip {
    name          = "default_gzip"
    content_types = ["text/html", "text/css"]
    extensions    = ["css", "js", "html"]
  }
```

Then run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_active_version` move to the new version
- the updated `content_types` reflected in the service configuration
- `service_2_*` outputs remain unchanged

### Configure Image Optimizer default settings

Once the bootstrap apply has enabled Image Optimizer on `service_1` via
`fastly_service_product_image_optimizer.service_1`, add the
`image_optimizer_default_settings` block to `fastly_service_cdn_auto.service_1`
in `main.tf`:

```hcl
  image_optimizer_default_settings {
    resize_filter = "lanczos3"
    webp          = false
    webp_quality  = 85
    jpeg_type     = "auto"
    jpeg_quality  = 85
    upscale       = false
    allow_video   = false
  }
```

Then run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_active_version` move to the new version
- `service_1_image_optimizer_default_settings` reflect the configured values

### Change Image Optimizer default settings

For example, change `resize_filter` from `lanczos3` to `bicubic` on `service_1`,
then run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_active_version` move to the new version
- `service_1_image_optimizer_default_settings` reflect the new value

### Remove Image Optimizer default settings

Delete the `image_optimizer_default_settings` block from `service_1` entirely
and run:

```bash
terraform apply
```

You should see:

- `service_1_managed_version` increase
- `service_1_image_optimizer_default_settings` become an empty list
- Image Optimizer default settings on the service reset back to Fastly's API
  defaults (rather than staying at their last-configured values)

## Notes

- This example is for the automatic compatibility resource family only.
- Do not mix `fastly_service_cdn_auto` with first-class explicit/default
  resources for the same Fastly service.
- This first cut is intended to mirror the current provider’s
  convenience-oriented lifecycle behavior for service, domain, and backend.
- At most one `image_optimizer_default_settings` block is supported per
  service, and it requires the Image Optimizer product to already be enabled
  on that service.
