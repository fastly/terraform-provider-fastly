# Orchestration Example (Terraform Actions)

This example demonstrates how to manage **multiple Fastly services** with the
**version-agnostic Fastly Terraform demo provider**, while performing **version
cloning and activation explicitly with Terraform Actions**.

The example provisions and manages:

- **Two Fastly services**
- A **shared backend definition** reused across both services
- A **service-specific backend** on only one service

## What this example shows

- Managing **multiple Fastly services** in one Terraform configuration
- Reusing shared configuration across services
- Attaching `fastly_service_domain` and `fastly_service_backend` resources to an explicit service version
- Using `data.fastly_service_version` to inspect current version state
- Explicit version cloning with Terraform Actions
- Explicit version activation with Terraform Actions

## Files

```text
examples/orchestration-actions/
  main.tf
  variables.tf
  outputs.tf
  terraform.tfvars
  modules/service/
  scripts/
    changed-services.sh
```

## Local engineer workflow

```bash
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json

terraform plan  -invoke=action.fastly_service_version_clone.service_1
terraform apply -invoke=action.fastly_service_version_clone.service_1
terraform plan  -invoke=action.fastly_service_version_clone.service_2
terraform apply -invoke=action.fastly_service_version_clone.service_2

# update terraform.tfvars version pins

# optional verification before opening the PR
terraform plan
```

## CI

```bash
terraform fmt -check -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json
```

## CD

Ideally, CD applies the reviewed `tfplan` artifact produced earlier in the workflow. If that plan artifact is not available in CD, generate a new plan first.

```bash
# if tfplan is already available in CD
terraform apply tfplan
terraform show -json tfplan > tfplan.json
./scripts/activate-changed-services.sh tfplan.json
```
