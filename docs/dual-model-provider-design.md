# Dual-Model Provider Design

## Purpose

This document defines the design of the dual-model Fastly Terraform provider.

The provider exposes two resource families:

1. an **explicit/default resource family** with first-class resources and
   caller-managed version lifecycle operations
2. an **automatic compatibility resource family** with nested configuration and
   automatic version lifecycle behavior

The goal is to make the provider model, resource naming, version lifecycle,
activation behavior, staging behavior, import behavior, and query behavior
explicit enough to review, implement, test, and document for users.

---

## Design goals

The dual-model provider design should provide the following properties:

- make the explicit workflow the default resource naming surface
- preserve current-provider-style convenience behavior for users who want
  automatic clone and activation
- provide first-class versioned resources for users who want explicit lifecycle
  control
- avoid provider-only `mode` fields
- avoid provider-only `type` fields
- make the selected model clear from the resource type
- support only two lifecycle workflows:
  - explicit / caller-managed lifecycle
  - automatic compatibility lifecycle
- make staging explicit-only
- support generated configuration for both resource families using the mechanism
  that fits each model

---

## Resource naming

The explicit workflow uses the clean/default resource names.

The automatic compatibility workflow uses the `_auto` suffix.

| Workflow | CDN service | Compute service |
| --- | --- | --- |
| Explicit/default | `fastly_service_cdn` | `fastly_service_compute` |
| Automatic compatibility | `fastly_service_cdn_auto` | `fastly_service_compute_auto` |

Shared first-class versioned resources also use clean names:

- `fastly_service_domain`
- `fastly_service_backend`

This avoids an `_explicit` suffix on most resources and makes the preferred
explicit workflow feel like the primary provider interface.

---

## Resource families

### Explicit/default resource family

The explicit/default family uses first-class resources.

For CDN services, the service resource is separate from the versioned resources
that belong to a service version:

```hcl
resource "fastly_service_cdn" "example" {
  name = "example"
}

resource "fastly_service_domain" "example" {
  service_id = fastly_service_cdn.example.id
  version    = var.service_version
  name       = "www.example.com"
}

resource "fastly_service_backend" "example" {
  service_id = fastly_service_cdn.example.id
  version    = var.service_version
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

For Compute services, the service resource follows the same explicit workflow:

```hcl
resource "fastly_service_compute" "example" {
  name = "example"
}

resource "fastly_service_domain" "example" {
  service_id = fastly_service_compute.example.id
  version    = var.service_version
  name       = "www.example.com"
}

resource "fastly_service_backend" "example" {
  service_id = fastly_service_compute.example.id
  version    = var.service_version
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

The explicit/default family does not clone or activate versions during normal
resource CRUD. The caller is responsible for selecting a writable version and
performing version lifecycle operations explicitly.

Shared explicit versioned resources, such as domains and backends, may support
both CDN and Compute services. Versioned resources that apply only to one service
kind must validate the target service kind before writing.

### Automatic compatibility resource family

The automatic compatibility family uses aggregate service resources with nested
versioned configuration.

For CDN services:

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

For Compute services:

```hcl
resource "fastly_service_compute_auto" "example" {
  domain {
    name = "www.example.com"
  }

  backend {
    name    = "origin"
    address = "origin.example.com"
    port    = 443
  }

  package {
    filename         = "${path.module}/pkg/package.tar.gz"
  }
}
```

Automatic compatibility service resources own the full service configuration for
the versioned resource types nested under them.

---

## No `mode` field

The provider does not expose a `mode` field.

A `mode` field would be provider-only metadata. It is not part of the Fastly API
and cannot be derived from remote state during import or query operations.

Instead, the resource type defines the lifecycle model:

- clean/default service and versioned resource names mean explicit /
  caller-managed lifecycle behavior
- `_auto` service resources mean automatic compatibility lifecycle behavior

---

## No `type` field

The provider does not expose a `type` field such as `delivery`, `vcl`, `wasm`,
`cdn`, or `compute`.

Service kind is represented by the resource type, not by provider-only metadata.

CDN service resources:

- `fastly_service_cdn`
- `fastly_service_cdn_auto`

Compute service resources:

- `fastly_service_compute`
- `fastly_service_compute_auto`

---

## Behavior matrix

Only two lifecycle workflows are supported.

| Resource family | Clone behavior | Version selection | Activation | Staging |
| --- | --- | --- | --- | --- |
| Explicit/default resources | Caller-managed | Caller-managed | Manual / explicit | Supported only here |
| Automatic compatibility resources | Provider-managed | Provider-managed | Automatic | Not supported |

The provider should not support intermediate variants:

- automatic compatibility + manual activation
- explicit/default resources + automatic activation

There is intentionally no `activate` or `stage` field on automatic compatibility
service resources. The automatic compatibility resource family always performs
automatic clone and activation during normal CRUD. Users who do not want that
behavior should use the explicit/default resource family instead.

---

## Automatic compatibility resource behavior

### Core rule

Automatic compatibility service resources are the automatic lifecycle surface.

They automatically:

1. create or select a writable service version
2. reconcile nested versioned resources into that version
3. validate the version
4. activate the version

### Bootstrap behavior

When an automatic compatibility service resource creates a new Fastly service,
Fastly service creation already creates version `1`.

The provider should use that version for initial nested versioned configuration.
It does not need to create an initial version itself.

Bootstrap flow:

```text
Create automatic compatibility service resource
  |
  v
Fastly creates service and editable version 1
  |
  v
Provider writes nested versioned resources to version 1
  |
  v
Provider validates version 1
  |
  v
Provider activates version 1
```

### Update behavior

When nested versioned configuration changes, the automatic compatibility service
resource selects one working version for the service update.

All changed nested versioned resources for that service are written to the same
working version.

The provider must not clone once per changed nested versioned resource.

Update flow:

```text
automatic compatibility service resource detects changed nested versioned configuration
  |
  v
select one working version for the service
  |
  v
reconcile all nested versioned resources to that version
  |
  v
validate selected version
  |
  v
activate selected version automatically
```

### Bootstrap version selection

When an automatic compatibility service resource creates a new Fastly service,
Fastly service creation already creates version `1`.

The provider uses version `1` directly for the initial nested versioned
configuration, then validates and activates it.

### Update version selection

When an automatic compatibility service resource updates an existing service, it
should use this selection order:

1. If the service has an active version, clone the active version.
2. If the service has no active version, clone the latest existing version.

This gives predictable behavior for both activated services and services that
have draft history but no active version.

### Compatibility flow diagram

```text
Service S update through automatic compatibility service resource
  |
  | active version exists?
  |---- yes ---> clone active version
  |
  |---- no ----> clone latest existing version
  v
working version selected for service S
  |
  v
all nested versioned resource changes for service S write to that same version
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
- Automatic activation is not optional in the automatic compatibility family.
- Staging is not supported in the automatic compatibility family.

---

## Explicit/default resource behavior

### Core rule

In the explicit/default family, versioned resources specify `version`.

The provider writes directly to the specified version.

The caller chooses the target version.

The provider checks target version mutability and rejects writes to active or
locked versions with a clear diagnostic.

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
All explicit/default versioned resources for service S write directly to vN
  |
  v
User or pipeline activates or stages vN explicitly
```

### Service-kind validation

Some first-class versioned resources can be used with both CDN and Compute
services. Other first-class versioned resources are service-kind specific.

The provider should validate the service kind before writing when a resource is
not supported by all service kinds.

Examples:

- domains and backends can be shared by CDN and Compute services
- CDN-only resources should reject Compute service IDs
- Compute-only resources should reject CDN service IDs

This validation usually cannot be guaranteed by `terraform validate`, because
`service_id` may be computed, imported, or come from remote state. The provider
should therefore return clear diagnostics during planning or CRUD when the
service kind is known.

### Compute package upload

Compute packages are associated with a specific service version. In the
explicit/default family, the service resource itself remains versionless, while
the package upload targets the caller-selected service version.

For that reason, package upload is not modeled as an attribute or nested block on
`fastly_service_compute`.

It should also not be modeled as a normal first-class managed resource unless the
Fastly API supports full resource lifecycle semantics, including delete. Compute
package upload is modeled as an explicit action because package upload is a
versioned operation and the API supports upload/read semantics, but not a
dedicated package delete operation.

A simple explicit/default Compute workflow can use `terraform_data` as the
stateful trigger and an action to upload the package during normal
`terraform apply`:

```hcl
variable "service_version" {
  type = number
}

resource "fastly_service_compute" "app" {
  name = "example-compute-service"
}

resource "fastly_service_domain" "app" {
  service_id = fastly_service_compute.app.id
  version    = var.service_version
  name       = "www.example.com"
}

resource "fastly_service_backend" "origin" {
  service_id = fastly_service_compute.app.id
  version    = var.service_version
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}

resource "terraform_data" "compute_package" {
  input = {
    service_id       = fastly_service_compute.app.id
    version          = var.service_version
    filename         = "${path.module}/pkg/package.tar.gz"
    source_code_hash = filebase64sha256("${path.module}/pkg/package.tar.gz")
  }

  lifecycle {
    action_trigger {
      action = action.fastly_service_compute_package_upload.this
      events = ["create", "update"]
    }
  }
}

action "fastly_service_compute_package_upload" "this" {
  config {
    service_id       = fastly_service_compute.app.id
    version          = var.service_version
    filename         = "${path.module}/pkg/package.tar.gz"
  }
}
```

With this pattern, users run a normal `terraform apply` to create or update the
Compute package on the pinned service version. The upload action is triggered
when `terraform_data.compute_package` is created or updated, for example because
the target version or `source_code_hash` changed.

Activation or staging remains a separate explicit lifecycle step.

### Notes

- In explicit/default mode, the provider does not choose the target version.
- The caller chooses the target version.
- The provider checks target version mutability and rejects writes to active or
  locked versions with a clear diagnostic.
- This is the controlled workflow for auditability, rollout control, and
  staging.

---

## Staging

Staging is supported only through the explicit/default resource family.

Automatic compatibility resources should not stage automatically.

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
automatic compatibility family’s automatic clone-and-activate lifecycle.

---

## Activation

### Automatic activation

Automatic activation is supported only by the automatic compatibility resource
family.

For service resources, that means:

- `fastly_service_cdn_auto`
- `fastly_service_compute_auto`

Automatic activation is best suited to:

- users who want behavior close to the current provider
- fire-and-forget workflows
- simple service updates where immediate production activation is expected

Automatic activation must not be combined with staging.

### Manual activation

Manual or explicit activation is supported only by the explicit/default resource
family.

It is better suited to:

- coordinated multi-service production rollouts
- pre-production validation followed by an explicit production promotion step
- service chaining workflows where related services may need to be activated in
  a specific order
- staging-first workflows
- CI/CD systems that want an explicit promotion step

---

## Do not mix resource families for one service

A single Fastly service should be managed through one resource family only:

- explicit/default resources
- automatic compatibility resources

Do not manage the same Fastly service with both an automatic compatibility
service resource and explicit/default first-class resources.

The provider may not be able to reliably detect mixed ownership in all cases,
especially across separate Terraform states or imported resources. Mixed
ownership can cause drift, conflicting writes, or unexpected version lifecycle
behavior.

---

## Import and query support

The two resource families use different configuration-generation workflows.

### Automatic compatibility family: generated configuration through import

The automatic compatibility family uses nested configuration.

Example:

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

Because automatic compatibility service resources own nested service
configuration as one aggregate resource, `terraform query` is not the right
mechanism for generating that configuration.

For the automatic compatibility family, generated configuration should come from
Terraform import:

```bash
terraform import fastly_service_cdn_auto.example <service_id>
```

During read/import, automatic compatibility service resources select the version
to read from using this order:

1. If the service has an active version, read from the active version.
2. If the service has no active version, read from the latest service version.

### Explicit/default family: generated configuration through query

The explicit/default family uses independent first-class resources.

Example:

```hcl
resource "fastly_service_cdn" "example" {
  name = "example"
}

resource "fastly_service_domain" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "www.example.com"
}

resource "fastly_service_backend" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

Because these resources are first-class, `terraform query` is the right mechanism
for discovery. When used with `-generate-config-out`, it can generate
configuration and import blocks for the explicit/default resource family.

For query-based discovery, the provider selects the version to read from using
this order:

1. If the service has an active version, read from the active version.
2. If the service has no active version, read from the latest service version.

A Fastly service is expected to have at least one version because service
creation creates version `1`.

When configuration is generated, the generated explicit/default resources include
the version number that was read.

If that version is active or locked, the generated configuration is still useful
for discovery. Before making changes with the explicit/default resource family,
the user must clone or select a writable version and update the generated
resources to target that writable version.

### Query flow diagram

```text
terraform query for explicit/default resources
  |
  v
inspect service S
  |
  | active version exists?
  |---- yes ---> read first-class versioned resources from active version
  |
  |---- no ----> read first-class versioned resources from latest version
  v
generate first-class resources with version pinned to the version read
  |
  v
read-only discovery; no version lifecycle operations
```
