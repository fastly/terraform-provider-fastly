# Orchestration Example (CLI / Scripts)

This example demonstrates how to manage **multiple Fastly services** with the
**version-agnostic Fastly Terraform demo provider**, while performing **version
cloning and activation explicitly with the Fastly CLI**.

The example provisions and manages:

- **Two Fastly services**
- A **shared backend definition** reused across both services
- A **service-specific backend** on only one service

## What this example shows

- Managing **multiple Fastly services** in one Terraform configuration
- Reusing shared configuration across services
- Attaching `fastly_service_domain` and `fastly_service_backend` resources to an explicit service version
- Explicit version cloning with the Fastly CLI
- Explicit version activation with the Fastly CLI
- No hidden cloning or activation during normal `terraform apply`

## Files

```text
examples/orchestration-cli/
  main.tf
  variables.tf
  outputs.tf
  terraform.tfvars
  modules/service/
  scripts/
    clone.sh
    activate.sh
    changed-services.sh
```

## Local engineer workflow

```bash
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/changed-services.sh tfplan.json

./scripts/clone.sh service-1 active
./scripts/clone.sh service-2 active

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
