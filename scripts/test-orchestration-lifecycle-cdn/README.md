# Orchestration Lifecycle Test Suite

This directory contains an end-to-end test suite for the Fastly Terraform provider's orchestration features, including resource lifecycle management and action invocations.

## Overview

The test suite validates the complete lifecycle of Fastly services managed via Terraform, including:

- Service creation (`fastly_service_cdn`)
- Domain attachment (`fastly_service_domain`)
- Backend configuration (`fastly_service_backend`)
- Version data sources (`data.fastly_service_version`)
- Resource updates
- Version cloning action (`fastly_service_version_clone`)
- Version activation action (`fastly_service_version_activate`)
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

## Running the Tests

From the repository root:

```bash
./scripts/test-orchestration-lifecycle/run.sh
```

Or from this directory:

```bash
./run.sh
```

## What the Test Does

1. **Build Provider**: Compiles the Terraform provider from source
2. **Setup Environment**: Creates a temporary test directory with Terraform configuration
3. **Initialize Terraform**: Runs `terraform init` and `terraform validate`
4. **Apply Initial Config**: Creates two test services with domains and backends
5. **Verify State**: Checks that resources were created correctly
6. **Test Updates**: Modifies a service comment to verify updates work
7. **Test Clone Action**: Invokes version clone actions for both services
8. **Test Activate Action**: Invokes version activate action
9. **Test Destruction**: Cleans up all resources via `terraform destroy`
10. **Cleanup**: Removes temporary test directory

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

## Troubleshooting

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
