# Orchestration Example (Latest-Version Convenience Workflow)

This example demonstrates a **convenience-oriented workflow** for the
version-agnostic Fastly Terraform demo provider. It uses the
`fastly_service_version` data source to target the **latest service version**
instead of an explicitly pinned version variable.

The example provisions and manages:

- **Two Fastly services**
- A **shared backend definition** reused across both services
- A **service-specific backend** on only one service

## Important caveats

This workflow is intentionally provided as an **optional example**, not as the
recommended workflow for the version-agnostic provider design.

It has important tradeoffs:

- it is **not recommended for CI/CD**
- it is **not recommended when multiple engineers or multiple pending PRs** may
  be active at the same time
- it depends on the latest version at the time Terraform plans and applies
- it is best suited only to a **careful, single-user, one-change-at-a-time**
  workflow

If you need stronger auditability, predictable review of exact version numbers,
or a CI/CD-friendly workflow, use the pinned-version examples instead:

- `examples/orchestration-cli`
- `examples/orchestration-actions`

## What this example shows

- how to use `data.fastly_service_version.*.latest_version` instead of explicit
  version pins in Terraform variables
- how version lifecycle can still remain explicit even when the version number
  is not pinned in configuration

## Files

```text
examples/orchestration-latest/
  main.tf
  variables.tf
  outputs.tf
  terraform.tfvars
  modules/service_config/
  scripts/
    clone.sh
    activate.sh
    changed-services.sh
    activate-changed-services.sh
```

## How this example works

The services are created as normal `fastly_service` resources.

The versioned resources use the version data source to reference the latest
version number for each service:

```hcl
data "fastly_service_version" "service_1" {
  service_id = fastly_service.service_1.id
}

module "service_1" {
  source          = "./modules/service_config"
  service_id      = fastly_service.service_1.id
  service_version = data.fastly_service_version.service_1.latest_version
  domain_name     = "www.service1.example.com"
  backends        = [...]
}
```

After a clone operation creates a new draft version, a normal `terraform apply`
will pick up that latest version automatically.

## One-time bootstrap

For a brand new environment, initialize and apply once to create the services and
initial versioned resources:

```bash
cd examples/orchestration-latest
terraform init
terraform apply
```

## Day-to-day engineer workflow

### 1. Update the Terraform configuration locally

Make the desired changes in HCL.

### 2. Inspect the local change

Use a local plan and the helper script to see which versioned resources and
services are affected.

```bash
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json
```

### 3. Clone the affected latest versions

Clone only the services that actually changed.

Examples:

```bash
./scripts/clone.sh service-1 latest
./scripts/clone.sh service-2 latest
```

Use `latest` consistently in this workflow. Because the configuration also
targets the latest version, the next `terraform apply` will use the newly cloned
draft version.

### 4. Apply the configuration changes

Do **not** apply the saved `tfplan` from step 2. That plan was generated before
the clone, so it may point at an older latest version.

Instead, run a fresh apply after cloning so Terraform can resolve the current
latest version again:

```bash
terraform apply
```

### 5. Activate only the changed services

Use the helper script with the pre-clone `tfplan.json` to activate only the
services that had versioned changes.

```bash
./scripts/activate-changed-services.sh tfplan.json
```

## Why this is not recommended for CI/CD

This workflow relies on resolving `latest_version` dynamically rather than
pinning an exact service version in configuration.

That means:

- plans can vary over time depending on external version state
- two engineers can see different results at different times
- a reviewed plan artifact cannot safely be reused after a clone
- multiple pending changes can become confusing or unsafe

For those reasons, this example is best treated as a convenience workflow for
careful manual use, not as the recommended model for automated pipelines.

## Recommended alternative

For CI/CD, reproducibility, and auditability, prefer the pinned-version
workflows:

- `examples/orchestration-cli`
- `examples/orchestration-actions`
