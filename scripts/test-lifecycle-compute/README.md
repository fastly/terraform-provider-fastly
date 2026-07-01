# Compute Service Lifecycle Test Suite

This directory contains an end-to-end test suite for the Fastly Terraform provider's Compute service features, including resource lifecycle management and action invocations.

## Overview

The test suite validates the complete lifecycle of Fastly Compute services managed via Terraform, including:

- Service creation (`fastly_service_compute`)
- Domain attachment (`fastly_service_domain`)
- Backend configuration (`fastly_service_backend`)
- ACL configuration (`fastly_service_acl`)
- Version data sources (`data.fastly_service_version`)
- Compute package uploads (`fastly_service_compute_package_upload` action)
- Version cloning action (`fastly_service_version_clone`)
- Version activation action (`fastly_service_version_activate`)
- Resource updates
- Resource destruction

## Files

- **`run.sh`** - Main test script
- **`main.tf`** - Terraform resources and actions configuration
- **`variables.tf`** - Input variable definitions
- **`outputs.tf`** - Output value definitions

## Prerequisites

1. **Fastly API Token**: Set the `FASTLY_API_TOKEN` environment variable
2. **Terraform**: Install Terraform CLI (tested with v1.15+)
3. **Go**: Install Go for building the provider
4. **jq**: Install jq for JSON processing
5. **Compute Package**: The test uses the package at `internal/acceptance_tests/fixtures/packages/valid.tar.gz`

## Running the Tests

From the repository root:

```bash
./scripts/test-lifecycle-compute/run.sh
```

Or from this directory:

```bash
cd scripts/test-lifecycle-compute
./run.sh
```

## What the Test Does

1. **Build Provider**: Compiles the Terraform provider from source
2. **Setup Environment**: Creates a temporary test directory with Terraform configuration
3. **Initialize Terraform**: Runs `terraform init` and `terraform validate`
4. **Apply Initial Config**: Creates two test Compute services with domains and backends
5. **Verify State**: Checks that resources were created correctly
6. **Test Package Upload**: Invokes `fastly_service_compute_package_upload` actions for both services
7. **Test Resource Updates**: Modifies a service comment while version 1 is still editable
8. **Test Clone Action**: Invokes version clone actions for both services (creates version 2)
9. **Test Activate Action**: Uploads package and activates version 2
10. **Test Version-Locked Writes**: Clones to version 3, adds new resources, uploads package, and activates
11. **Test Destruction**: Cleans up all resources via `terraform destroy`
12. **Cleanup**: Removes temporary test directory

## Resource Cleanup

The test includes robust cleanup logic using Terraform:

- Removes version-locked resources from Terraform state (backends, domains, and ACLs pinned to old versions)
- Runs `terraform destroy` to delete services
- The `force_destroy = true` flag ensures all versions are deleted automatically
- Verifies services were successfully deleted via the Fastly API
- Emergency cleanup on test failures via API calls

## Compute-Specific Features

This test suite specifically validates Compute service functionality:

- **Package Upload Action**: Tests the `fastly_service_compute_package_upload` action which is unique to Compute services
- **Version Activation Requirements**: Compute versions require a package to be uploaded before they can be activated

## Troubleshooting

### Package Not Found

If you see "Compute package not found" error, ensure the package exists at:
```
internal/acceptance_tests/fixtures/packages/valid.tar.gz
```

### Version Activation Failures

Compute versions require a package to be uploaded before activation. The test handles this by uploading the package to version 2 before attempting to activate it.

### Version Locked Resources

After cloning versions, backends, domains, and ACLs remain pinned to version 1 in Terraform state, which becomes locked after activation. The script handles this by removing these resources from state before running `terraform destroy`. This allows the service deletion (with `force_destroy = true`) to clean up all versions and child resources automatically.

### API Rate Limiting

If tests fail due to rate limiting, wait a few minutes before re-running the test suite.

## Resources Created

The test creates:
- 2 Fastly Compute services
- 2 domain attachments (one per service)
- 3 backend configurations (1 shared, 2 unique)
- 2 ACL configurations (one per service)
- Compute package uploads to multiple versions

All resources are tagged with the process ID for uniqueness and are destroyed during cleanup.

## Extending the Tests

To add new test scenarios:

1. Add resources to `main.tf`
2. Add variables to `variables.tf` if needed
3. Add outputs to `outputs.tf` if needed
4. Add test functions to `run.sh` following the naming convention `test_*`
5. Call the new test function from `main()` in `run.sh`
