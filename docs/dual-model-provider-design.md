# Dual-Model Provider Design

## Purpose

This document defines the design of the dual-model Fastly Terraform provider.

The provider exposes two separate resource families:

1. a **compatibility resource family** with nested configuration and automatic
   version lifecycle behavior
2. an **explicit resource family** with first-class versioned resources and
   explicit version lifecycle operations

The goal is to make the provider model, resource naming, version lifecycle,
activation behavior, staging behavior, import behavior, and query behavior
explicit enough to review, implement, test, and document for users.

---

## Design goals

The dual-model provider design should provide the following properties:

- preserve current-provider-style convenience behavior for users who want
  automatic clone and activation
- provide first-class versioned resources for users who want explicit lifecycle
  control
- avoid any provider-only fields
- make the selected model clear from the resource type
- support only two lifecycle workflows:
  - compatibility / automatic lifecycle
  - explicit / caller-managed lifecycle
- make staging explicit-only
- support generated configuration for both resource families using the mechanism
  that fits each model

---

## Resource families

### Compatibility resource family

The compatibility family uses an aggregate service resource with nested
versioned configuration.

Current VCL resource:

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

`fastly_service_vcl` owns the full service configuration for the resource types
nested under it.

Future Compute support should follow the same naming pattern:

- `fastly_service_compute`

### Explicit resource family

The explicit family uses first-class resources.

Current VCL resources:

```hcl
resource "fastly_service_vcl_explicit" "example" {
  name = "example"
}

resource "fastly_service_domain_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = var.service_version
  name       = "www.example.com"
}

resource "fastly_service_backend_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = var.service_version
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

The explicit family does not clone or activate versions during normal resource
CRUD. The caller is responsible for selecting a writable version and performing
version lifecycle operations explicitly.

Future Compute support should follow the same naming pattern:

- `fastly_service_compute_explicit`

---

## No `mode` field

The provider does not expose a `mode` field.

A `mode` field would be provider-only metadata. It is not part of the Fastly API
and cannot be derived from remote state during import or query operations.

Instead, the resource type defines the lifecycle model:

- `fastly_service_vcl` means compatibility / automatic lifecycle behavior
- `*_explicit` resources mean explicit / caller-managed lifecycle behavior

---

## No `type` field

The provider does not expose a `type` field such as `delivery`, `vcl`, or
`compute`.

Service kind should be represented by the resource type, not by provider-only
metadata.

Current VCL service resources:

- `fastly_service_vcl`
- `fastly_service_vcl_explicit`

Future Compute service resources should be separate resources:

- `fastly_service_compute`
- `fastly_service_compute_explicit`

---

## Behavior matrix

Only two lifecycle workflows are supported.

| Resource family                     | Clone behavior   | Version selection | Activation        | Staging             |
| ----------------------------------- | ---------------- | ----------------- | ----------------- | ------------------- |
| Compatibility: `fastly_service_vcl` | Provider-managed | Provider-managed  | Automatic         | Not supported       |
| Explicit: `*_explicit`              | Caller-managed   | Caller-managed    | Manual / explicit | Supported only here |

The provider should not support intermediate variants:

- compatibility + manual activation
- explicit resources + automatic activation

There is intentionally no `activate` field on `fastly_service_vcl`. The
compatibility resource family always performs automatic clone and activation
during normal CRUD. Users who do not want that behavior should use the explicit
resource family instead.

---

## Compatibility resource behavior: `fastly_service_vcl`

### Core rule

`fastly_service_vcl` is the compatibility surface.

It automatically:

1. creates or selects a writable service version
2. reconciles nested configuration into that version
3. validates the version
4. activates the version

### Bootstrap behavior

When `fastly_service_vcl` creates a new Fastly service, Fastly service creation
already creates version `1`.

The provider should use that version for initial nested configuration. It does
not need to create an initial version itself.

Bootstrap flow:

```text
Create fastly_service_vcl
  |
  v
Fastly creates service and editable version 1
  |
  v
Provider writes nested domains/backends to version 1
  |
  v
Provider validates version 1
  |
  v
Provider activates version 1
```

### Update behavior

When nested versioned configuration changes, `fastly_service_vcl` selects one
working version for the service update.

All changed nested configuration for that service is written to the same working
version.

The provider must not clone once per nested block.

Update flow:

```text
fastly_service_vcl detects changed nested configuration
  |
  v
select one working version for the service
  |
  v
reconcile all nested domains to that version
  |
  v
reconcile all nested backends to that version
  |
  v
validate selected version
  |
  v
activate selected version automatically
```

### Bootstrap version selection

When `fastly_service_vcl` creates a new Fastly service, Fastly service creation
already creates version `1`.

The provider uses version `1` directly for the initial nested configuration,
then validates and activates it.

### Update version selection

When `fastly_service_vcl` updates an existing service, it should use this
selection order:

1. If the service has an active version, clone the active version.
2. If the service has no active version, clone the latest existing version.

This gives predictable behavior for both activated services and services that
have draft history but no active version.

### Compatibility flow diagram

```text
Service S update through fastly_service_vcl
  |
  | active version exists?
  |---- yes ---> clone active version
  |
  |---- no ----> clone latest existing version
  v
working version selected for service S
  |
  v
all nested domain/backend changes for service S write to that same version
  |
  v
validate selected version
  |
  v
activate selected version automatically
```

### Notes

- The key property is **one working version per changed service update**.
- For newly created services, bootstrap happens through service creation, which
  already creates version `1`.
- Automatic activation is not optional in the compatibility family.
- Staging is not supported in the compatibility family.

---

## Explicit resource behavior: `*_explicit` resources

### Core rule

In the explicit family, versioned resources specify `version`.

The provider writes directly to the specified version.

The caller is responsible for ensuring that the version is writable.

Clone timing, version selection, activation, and staging are controlled by the
caller or by external automation.

### Explicit flow diagram

```text
Service S

User or pipeline clones or selects writable version vN
  |
  v
Terraform configuration pins version = vN
  |
  v
Terraform apply
  |
  v
All explicit versioned resources for service S write directly to vN
  |
  v
User or pipeline activates or stages vN explicitly
```

### Notes

- In explicit mode, the provider does not choose the target version.
- The caller is responsible for ensuring the chosen version is writable.
- The provider may check version mutability and reject writes to active or
  locked versions.
- This is the controlled workflow for auditability, rollout control, and
  staging.

---

## Staging

Staging is supported only through the explicit resource family.

Compatibility resources should not stage automatically.

### Staging flow diagram

```text
Service S
production active version = vP

User or pipeline clones or selects writable version vN
  |
  v
Terraform configuration pins version = vN
  |
  v
Terraform apply writes changes to vN
  |
  v
User or pipeline stages vN
  |
  v
Test and evaluate vN on staging
  |
  +--> promote later? yes -> activate vN to production
  |
  +--> continue testing? -> update vN and stage again
```

### Why staging is explicit-only

Staging is a separate lifecycle path that requires testing and promotion
decisions.

It also has important Fastly-specific constraints, including:

- a service must already have at least one production version before staging can
  be used
- domains and TLS changes cannot be staged
- service chaining is not supported in staging
- versionless settings may affect both staging and production

Because staging requires explicit decisions, it is not a good fit for the
compatibility family’s automatic clone-and-activate lifecycle.

---

## Activation

### Automatic activation

Automatic activation is supported only by the compatibility resource family.

For the current VCL implementation, that means:

- `fastly_service_vcl`

Automatic activation is best suited to:

- users who want behavior close to the current provider
- fire-and-forget workflows
- simple service updates where immediate production activation is expected

Automatic activation must not be combined with staging.

### Manual activation

Manual or explicit activation is supported only by the explicit resource family.

It is better suited to:

- controlled rollout order
- pre-production validation
- service chaining
- staging-first workflows
- CI/CD systems that want an explicit promotion step

---

## Import and query support

The two resource families use different configuration-generation workflows.

### Compatibility family: generated configuration through import

The compatibility family uses nested configuration.

Example:

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

Because `fastly_service_vcl` owns the nested service configuration as one
aggregate resource, `terraform query` is not the right mechanism for generating
that configuration.

For the compatibility family, generated configuration should come from Terraform
import:

```bash
terraform import fastly_service_vcl.example <service_id>
```

During read/import, `fastly_service_vcl` selects the version to read from using
this order:

1. If the service has an active version, read from the active version.
2. If the service has no active version, read from the latest service version.

### Explicit family: generated configuration through query

The explicit family uses independent first-class resources.

Example:

```hcl
resource "fastly_service_vcl_explicit" "example" {
  name = "example"
}

resource "fastly_service_domain_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = 1
  name       = "www.example.com"
}

resource "fastly_service_backend_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = 1
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

Because these resources are first-class, `terraform query` is the right mechanism
for discovery and generated configuration.

For query-based discovery, the provider selects the version to read from using
this order:

1. If the service has an active version, read from the active version.
2. If the service has no active version, read from the latest service version.

A Fastly service is expected to have at least one version because service
creation creates version `1`.

Generated explicit resources include the version number that was read.

If that version is active or locked, the generated configuration is still useful
for discovery. Before making changes with the explicit resource family, the user
must clone or select a writable version and update the generated resources to
target that writable version.

### Query flow diagram

```text
terraform query for explicit resources
  |
  v
inspect service S
  |
  | active version exists?
  |---- yes ---> read domains/backends from active version
  |
  |---- no ----> read domains/backends from latest version
  v
generate *_explicit resources with version pinned to the version read
  |
  v
no clone, no activation, no staging, no mutation
```
