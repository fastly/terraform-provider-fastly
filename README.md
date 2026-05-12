# Fastly Terraform Provider Dual-Model Rewrite

This branch is the development branch for the Fastly Terraform provider rewrite
using the Terraform Plugin Framework.

The rewrite is developed on a long-lived feature branch:

```text
framework-rewrite-dual-model
```

Feature pull requests for the rewrite should target this branch while the new
provider implementation is in progress. After the rewrite is stable, this branch
is intended to be merged into `main` and released as a new major version.

## Design overview

The rewrite uses a **dual-model provider design** with two separate resource
families:

- a **compatibility resource family** for users who want current-provider-style
  nested resources with automatic clone and activation behavior
- an **explicit resource family** for users who want first-class versioned
  resources and explicit version lifecycle operations

The user chooses the model through the resource type.

## Resource families

### Compatibility family

```hcl
resource "fastly_service_vcl" "example" {
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

The compatibility family owns nested configuration and performs automatic
version lifecycle handling.

### Explicit family

```hcl
resource "fastly_service_vcl_explicit" "example" {
  name = "example"
}

resource "fastly_service_domain_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = var.service_version
  name       = "www.example.com"
}
```

The explicit family uses first-class resources. Version cloning, activation, and
staging are handled explicitly by the caller or workflow automation.

## Design documents

For the full design, see:

- [Dual-Model Provider Design](docs/dual-model-provider-design.md)
- [Terraform Query Support](docs/terraform-query.md)

## Examples

The examples compare the two resource families by managing the same basic Fastly
service configuration:

- `examples/orchestration-compat-vcl`
- `examples/orchestration-explicit-actions`
- `examples/orchestration-explicit-cli`
- `examples/orchestration-explicit-latest-cli`
- `examples/terraform-query-import`

## Important design rule

One Fastly service should be managed through **one resource family only**.

Do not manage the same Fastly service with both:

- `fastly_service_vcl`
- `*_explicit` resources
