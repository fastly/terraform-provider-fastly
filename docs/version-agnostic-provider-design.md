# A version-agnostic redesign for the Fastly Terraform Provider

We are proposing a redesign of the Fastly Terraform Provider to better align Terraform with Fastly’s versioned configuration model.

This post shares the proposed direction, why we think it improves the user experience, and how it would work in practice.

## Why change the provider?

Fastly includes both versioned and versionless resources. Many important configuration objects, such as backends, VCL, and logging endpoints, are tied to a specific service version, while other resources are not.

Terraform, on the other hand, is built around managing long-lived resources with predictable CRUD behavior.

That creates a recurring mismatch in the current provider for versioned service configuration, while versionless resources are not affected in the same way and do not require cloning or activation workflows.

That versioning mismatch creates a few recurring problems in the current provider:

- updating a resource can implicitly clone or activate a service version
- plans can be noisy and hard to review
- importing existing services can be awkward
- shared infrastructure is harder to model cleanly

The goal of this redesign is to make Fastly version management explicit for versioned resources, while keeping Terraform focused on declarative configuration and preserving the existing behavior of versionless resources.

## Core idea

In the proposed design:

- Terraform manages configuration only
- service version lifecycle is explicit
- versioned Fastly entities become first-class Terraform resources
- Terraform never implicitly clones or activates a service version

Instead of nesting many versioned objects under one large service resource, the provider uses flattened resources such as:

```hcl
resource "fastly_service" "app" {
  name = "example-app"
  type = "delivery"
}

resource "fastly_service_domain" "app" {
  service_id = fastly_service.app.id
  version    = var.service_version
  name       = "www.example.com"
}

resource "fastly_service_backend" "origin" {
  service_id = fastly_service.app.id
  version    = var.service_version
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

This keeps version pinning explicit and avoids hidden side effects during `terraform apply`.

## Version lifecycle stays explicit

Under this design, Terraform resources never clone or activate versions automatically.

Teams can manage version lifecycle in at least two ways:

- with the Fastly CLI
- with explicit Terraform Actions
- with their own tooling

The workflows below are presented as two supported options, not in recommendation order.

### Fastly CLI workflow

```text
Active version (v1)
  |
  | fastly service version clone --version=active
  v
Draft version (v2)
  |
  | update Terraform input
  | service_version = 2
  v
terraform apply
  |
  v
Updated draft version (v2)
  |
  | fastly service version activate --version=2
  v
Active version (v2)
```

### Terraform Actions workflow

```text
Active version (v1)
  |
  | terraform apply -invoke=fastly_service_version_clone
  v
Draft version (v2)
  |
  | update Terraform input
  | service_version = 2
  v
terraform apply
  |
  v
Updated draft version (v2)
  |
  | terraform apply -invoke=fastly_service_version_activate
  v
Active version (v2)
```

In both cases, cloning and activation are explicit lifecycle steps, not side effects of normal resource updates. After cloning, Terraform still requires an explicit version pin update before configuration changes are applied to the new draft version.

## Version introspection via data sources

To reduce guesswork, the provider can expose a version data source:

```hcl
data "fastly_service_version" "current" {
  service_id = fastly_service.app.id
}
```

This can expose:

- `latest_version`
- `active_version`
- `staging_version`
- `locked_versions`
- structured per-version metadata

For example, it can be used to provide the currently active production version number to the `fastly_service_version_clone` action:

```hcl
action "fastly_service_version_clone" "app" {
  config {
    service_id = fastly_service.app.id
    version    = data.fastly_service_version.current.active_version
  }
}
```

## Version mutability safety

Before modifying any versioned resource, the provider checks whether the targeted service version is editable.

If the version is active or locked, Terraform fails early with a clear error explaining the next step, such as cloning the version and pinning Terraform to the new draft version.

This improves safety and makes failures easier to understand.

## A real-world example

To make this concrete, the repo examples demonstrate a multi-service orchestration workflow:

- multiple Fastly services in one Terraform configuration
- shared configuration reuse across services
- explicit version cloning before review
- explicit version pin updates in Terraform inputs
- CI for `fmt`, `validate`, `plan`, and reviewer visibility
- CD for `apply`
- explicit activation after apply

The examples are provided in two variants:

- [`examples/orchestration-cli`](https://github.com/fastly/terraform-provider-fastly/tree/version-agnostic-design/examples/orchestration-cli)
- [`examples/orchestration-actions`](https://github.com/fastly/terraform-provider-fastly/tree/version-agnostic-design/examples/orchestration-actions)

Both show the same provider model with different lifecycle mechanisms.

## Recommended workflow

A typical workflow looks like this:

1. engineer updates HCL locally
2. engineer runs `fmt`, `validate`, and `plan` locally
3. engineer optionally exports the plan as JSON and uses helper tooling to identify affected versioned changes and services
4. engineer clones affected active versions
5. engineer updates version pins
6. engineer optionally runs a verification plan locally
7. engineer opens a pull request
8. CI runs `fmt`, `validate`, and `plan`, and may also run helper tooling that summarizes affected versioned changes and services so reviewers can verify the required version pin updates
9. CD runs `apply`
10. activation is performed only for services that actually changed

The key point is that cloning happens before review, and the updated version pins are part of the reviewed change.

## Feedback welcome

We are sharing this design to gather feedback before moving further.

The questions we are especially interested in are:

- does explicit version pinning feel clear and predictable?
- would your team prefer CLI-driven or Terraform Action-driven lifecycle steps?
- what example scenarios would be most useful to see documented?
