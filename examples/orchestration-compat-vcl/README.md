# Orchestration Example (`fastly_service_vcl` compatibility surface)

This example is the first orchestration-style example for the compatibility
surface in the dual-model Fastly Terraform provider. It uses `fastly_service_vcl` with nested
`domain` and `backend` blocks.

The example provisions and manages:

- two Fastly VCL services
- one shared backend reused across both services
- one service-specific backend on only service 1
- one domain per service

## What this example is for

This example is intended to test the current-provider-style behavior in the new
dual-model Fastly Terraform provider:

- one service resource owns the whole service configuration
- nested domain/backend changes are reconciled at the service level
- one draft version is cloned per changed service update
- the provider validates and activates automatically

## Files

```text
examples/orchestration-compat-vcl/
  main.tf
  variables.tf
  terraform.tfvars
  outputs.tf
  README.md
```

## One-time bootstrap

For a brand new environment:

```bash
cd examples/orchestration-compat-vcl
terraform init
terraform apply
```

Expected bootstrap behavior:

1. Terraform creates both `fastly_service_vcl` services.
2. Fastly creates editable version `1` for each new service.
3. The provider reconciles the nested `domain` and `backend` blocks to version `1`.
4. The provider validates and activates version `1`.

## Day-to-day workflow

Make configuration changes in `main.tf`, then run:

```bash
terraform fmt -recursive
terraform validate
terraform plan
terraform apply
```

Expected update behavior for a changed service:

1. the provider picks the current tracked service version as the source
2. it clones one new draft version for that service
3. it reconciles all nested domains to that draft
4. it reconciles all nested backends to that draft
5. it validates the draft
6. it activates the new version automatically

If only `service_1` changes, only `service_1` should get a new cloned and active
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

## Notes

- This example is for the compatibility surface only.
- Do not mix `fastly_service_vcl` with the explicit version-agnostic resources
  for the same Fastly service.
- This first cut is intended to mirror the current provider’s
  convenience-oriented lifecycle behavior for service, domain, and backend.
