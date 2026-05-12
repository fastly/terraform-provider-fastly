# Orchestration Example (Terraform Actions)

This example demonstrates how to manage **multiple Fastly services** with the
**dual-model Fastly Terraform provider rewrite explicit surface**, while
performing version cloning and activation explicitly with Terraform Actions.

The example provisions and manages:

- **Two Fastly services**
- A **shared backend definition** reused across both services
- A **service-specific backend** on only one service

## What this example shows

- Managing **multiple Fastly services** in one Terraform configuration
- Reusing shared configuration across services
- Attaching `fastly_service_domain_explicit` and
  `fastly_service_backend_explicit` resources to an explicit service version
- Using `data.fastly_service_version` to inspect current version state
- Explicit version cloning with Terraform Actions
- Explicit version activation and staging with Terraform Actions

## Files

```text
examples/orchestration-explicit-actions/
  main.tf
  variables.tf
  outputs.tf
  terraform.tfvars.example
  modules/service/
  scripts/
    changed-services.sh
```

## Setup

Copy the example variables file and fill in the values for your environment:

```bash
cp terraform.tfvars.example terraform.tfvars
```

## Local engineer workflow

```bash
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json
```

Clone only the services that changed. For example:

```bash
terraform apply -invoke=action.fastly_service_version_clone.service_1
terraform apply -invoke=action.fastly_service_version_clone.service_2
```

Then update the version pins in `terraform.tfvars` to point at the cloned
writable versions and run a normal plan/apply:

```bash
terraform plan
terraform apply
```

Activate or stage the changed services explicitly with Terraform Actions. For
example:

```bash
terraform apply -invoke=action.fastly_service_version_activate.service_1_prod
terraform apply -invoke=action.fastly_service_version_activate.service_2_prod
```

Optional staging invocations are also defined:

```bash
terraform apply -invoke=action.fastly_service_version_activate.service_1_staging
terraform apply -invoke=action.fastly_service_version_activate.service_2_staging
```

## CI

```bash
terraform fmt -check -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json
```

CI should report which services changed. Clone, version pin updates, apply, and
activation should happen through an explicit release workflow.

## Notes

- The explicit surface does not clone or activate during normal resource CRUD.
- Terraform Actions are invoked deliberately for clone, activation, and staging.
- This example intentionally avoids helper scripts for clone and activation so
  the lifecycle operations remain Terraform-native.
