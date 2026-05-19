# Compute Explicit Package Example

This example demonstrates a minimal explicit-family workflow for a Fastly
Compute service.

It manages:

- one `fastly_service_compute_explicit` service
- one explicit domain
- one explicit backend
- one Compute package upload action
- one explicit activation action

The Compute package upload is triggered during a normal `terraform apply` by a
`terraform_data` resource and `lifecycle.action_trigger`.

## Why package upload is an action

A Compute package belongs to a specific service version:

```text
service_id + version
```

The Fastly API supports reading and uploading a package for a service version,
but it does not expose a package delete operation. Because of that, this example
models package upload as an explicit Terraform Action instead of a managed
resource.

The action is attached to `terraform_data.compute_package`, which gives
Terraform a stateful object to diff. When the package path, target version, or
package hash changes, Terraform updates `terraform_data.compute_package` and the
package upload action runs.

## Files

```text
examples/compute-explicit-package/
  main.tf
  variables.tf
  outputs.tf
  terraform.tfvars.example
  pkg/
```

## Prepare a Compute package

Build or copy a valid Fastly Compute package to:

```text
examples/compute-explicit-package/pkg/package.tar.gz
```

The `pkg/` directory is intentionally empty except for `.gitkeep`.

## One-time setup

```bash
cd examples/compute-explicit-package
cp terraform.tfvars.example terraform.tfvars
terraform init
```

Edit `terraform.tfvars` and set:

- `fastly_api_key`
- `service_name`
- `domain_name`
- `backend`
- `package_filename`

For a brand-new explicit Compute service, `service_version = 1` is usually the
right initial value because Fastly creates version `1` when the service is
created.

## Create workflow

Run:

```bash
terraform apply
```

This creates the Compute service, writes the explicit domain/backend resources
to `var.service_version`, and uploads the Compute package to that same version.

No `terraform apply -invoke=...` is needed for the package upload. The upload is
triggered by the `terraform_data.compute_package` lifecycle action trigger.

Activation remains explicit. After reviewing the applied version, activate it
with:

```bash
terraform apply -invoke=action.fastly_service_version_activate.production
```

## Update workflow

For a package or configuration update:

1. clone or select a writable service version
2. update `service_version` in `terraform.tfvars`
3. build or copy the new package to `pkg/package.tar.gz`
4. run `terraform apply`

Terraform detects changes through:

- `var.service_version`
- `var.package_filename`
- `filebase64sha256(local.package_path)`

When `terraform_data.compute_package` changes, the package upload action runs
during the normal apply.

After validation/review, activate or stage the version explicitly.
