# Compute Service Orchestration Lifecycle Test Suite

This directory contains an end-to-end test suite for the Fastly Terraform provider's Compute service orchestration features, including resource lifecycle management and action invocations.

## Overview

The test suite validates the complete lifecycle of Fastly Compute services managed via Terraform, including:

- Service creation (`fastly_service_compute`)
- Domain attachment (`fastly_service_domain`)
- Backend configuration (`fastly_service_backend`)
- Version data sources (`data.fastly_service_version`)
- Compute package uploads (`fastly_compute_package_upload` action)
- Version cloning action (`fastly_service_version_clone`)
- Version activation action (`fastly_service_version_activate`)
- Resource updates
- Resource destruction

## Files

- **`run.sh`** - Main test orchestration script
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
./scripts/test-orchestration-lifecycle-compute/run.sh
```

Or from this directory:

```bash
./run.sh
```

## What the Test Does

1. **Build Provider**: Compiles the Terraform provider from source
2. **Setup Environment**: Creates a temporary test directory with Terraform configuration
3. **Initialize Terraform**: Runs `terraform init` and `terraform validate`
4. **Apply Initial Config**: Creates two test Compute services with domains and backends
5. **Verify State**: Checks that resources were created correctly
6. **Test Package Upload**: Invokes `fastly_compute_package_upload` actions for both services
7. **Test Clone Action**: Invokes version clone actions for both services
8. **Test Package Upload to v2**: Uploads package to the newly cloned version 2
9. **Test Activate Action**: Invokes version activate action to activate version 2
10. **Test Updates**: Modifies a service comment to verify updates work
11. **Test Destruction**: Cleans up all resources via `terraform destroy`
12. **Cleanup**: Removes temporary test directory

## Compute-Specific Features

This test suite specifically validates Compute service functionality:

- **Package Upload Action**: Tests the `fastly_compute_package_upload` action which is unique to Compute services
- **Version Activation Requirements**: Compute versions require a package to be uploaded before they can be activated
- **Service Type**: All services created use `fastly_service_compute` instead of `fastly_service_cdn`

## Resource Cleanup

The test includes robust cleanup logic using Terraform:

- Removes version-locked resources from Terraform state (backends and domains pinned to old versions)
- Runs `terraform destroy` to delete services
- The `force_destroy = true` flag ensures all versions are deleted automatically
- Verifies services were successfully deleted via the Fastly API
- Emergency cleanup on test failures via API calls

## Test Output

The script provides colored output with clear status indicators:

- 🔵 **[INFO]** - Informational messages
- 🟢 **[SUCCESS]** - Successful operations
- 🟡 **[WARNING]** - Non-critical warnings
- 🔴 **[ERROR]** - Critical errors

## Exit Codes

- `0` - All tests passed
- `1` - Test failure or error

## Differences from CDN Test Suite

This Compute test suite differs from the CDN lifecycle tests (`../test-orchestration-lifecycle/`) in:

1. Uses `fastly_service_compute` instead of `fastly_service_cdn`
2. Includes `fastly_compute_package_upload` action testing
3. Requires package upload before version activation
4. Tests Compute-specific workflows and constraints

## Troubleshooting

### Package Not Found

If you see "Compute package not found" error, ensure the package exists at:
```
internal/acceptance_tests/fixtures/packages/valid.tar.gz
```

### Version Activation Failures

Compute versions require a package to be uploaded before activation. The test handles this by uploading the package to version 2 before attempting to activate it.

### Version Locked Resources

After cloning versions, backends and domains remain pinned to version 1 in Terraform state, which becomes locked after activation. The script handles this by removing these resources from state before running `terraform destroy`. This allows the service deletion (with `force_destroy = true`) to clean up all versions and child resources automatically.

### Provider Override Warnings

Warnings about "provider development overrides" are expected when testing a locally-built provider. These can be safely ignored.

### API Rate Limiting

If tests fail due to rate limiting, wait a few minutes before re-running the test suite.

## Extending the Tests

To add new test scenarios:

1. Add resources to `main.tf`
2. Add variables to `variables.tf` if needed
3. Add outputs to `outputs.tf` if needed
4. Add test functions to `run.sh` following the naming convention `test_*`
5. Call the new test function from `main()` in `run.sh`

## Coverage

This test suite covers the following TODO items:

- ✅ Create basic CRUD tests for `fastly_service_compute`
- ✅ Test version cloning behaviors (Compute)
- ✅ Test version activation behaviors (Compute)
- ✅ Test resource writes against pinned versions (Compute)
- ✅ Test `fastly_compute_package_upload` action
- ✅ Service version data source (active and latest version queries)
