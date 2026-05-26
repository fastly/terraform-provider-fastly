# Fastly Terraform Provider Dual-Model Rewrite

This branch is the clean development branch for the Fastly Terraform provider
rewrite using the Terraform Plugin Framework.

The rewrite is developed on the orphan branch:

```text
dual-model-framework-rewrite
```

Feature pull requests for the rewrite should target this branch while the new
provider implementation is in progress.

The current `main` branch remains the existing provider line while the rewrite is
being developed. Once the rewrite is complete and stable, this branch is intended
to become the new main development line for the next major version of the Fastly
Terraform provider.

## Design overview

The rewrite uses a **dual-model provider design** with two separate resource
families:

- an **automatic compatibility resource family** for users who want
  current-provider-style nested resources with automatic clone and activation
  behavior
- an **explicit/default resource family** for users who want first-class
  versioned resources and explicit version lifecycle operations

The user chooses the model through the resource type.

The explicit/default resource family uses the clean resource names. The automatic
compatibility resource family uses the `_auto` suffix.

## Resource families

### Automatic compatibility family

```hcl
resource "fastly_service_cdn_auto" "example" {
  domain {
    name = "www.example.com"
  }

  backend {
    name    = "origin"
    address = "origin.example.com"
    port    = 443
  }
}
```

The automatic compatibility family owns nested configuration and performs
automatic version lifecycle handling.

Compatibility service resources:

- `fastly_service_cdn_auto`
- `fastly_service_compute_auto`

### Explicit/default family

```hcl
resource "fastly_service_cdn" "example" {
  name = "example"
}

resource "fastly_service_domain" "example" {
  service_id = fastly_service_cdn.example.id
  version    = var.service_version
  name       = "www.example.com"
}
```

The explicit/default family uses first-class resources. Version cloning,
activation, and staging are handled explicitly by the caller or workflow
automation.

Explicit/default service resources:

- `fastly_service_cdn`
- `fastly_service_compute`

Shared versioned resources include:

- `fastly_service_domain`
- `fastly_service_backend`

## Design documents

For the full design, see:

- [Dual-Model Provider Design](docs/dual-model-provider-design.md)
- [Terraform Query Support](docs/terraform-query.md)

## Examples

The examples compare the two resource families by managing the same basic Fastly
service configuration:

- `examples/orchestration-cdn-auto`
- `examples/orchestration-explicit-actions`
- `examples/orchestration-explicit-cli`
- `examples/orchestration-explicit-latest-cli`
- `examples/compute-explicit-package`
- `examples/terraform-query-import`

## Important design rule

One Fastly service should be managed through **one resource family only**.

Do not manage the same Fastly service with both:

- an automatic compatibility service resource such as `fastly_service_cdn_auto`
  or `fastly_service_compute_auto`
- explicit/default resources such as `fastly_service_cdn`,
  `fastly_service_compute`, `fastly_service_domain`, or
  `fastly_service_backend`

## Contributing to the rewrite

Rewrite work should happen through focused feature branches and pull requests
targeting `dual-model-framework-rewrite`.

Suggested workflow:

1. branch from `dual-model-framework-rewrite`
2. implement one focused part of the rewrite
3. open a pull request back into `dual-model-framework-rewrite`
4. repeat until the rewrite is complete and stable

Do not target rewrite pull requests at the current `main` branch.
