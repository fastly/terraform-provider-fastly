# Test Scripts

## test-orchestration-lifecycle/

A comprehensive integration test suite that validates the full orchestration lifecycle for Fastly services managed via Terraform.

### Directory Structure

```
test-orchestration-lifecycle/
├── README.md        # Detailed documentation
├── run.sh          # Main test orchestration script
├── main.tf         # Terraform resources and actions
├── variables.tf    # Input variable definitions
└── outputs.tf      # Output value definitions
```

### What it tests

✅ **Provider Build**
- Compiles the Terraform provider from source
- Configures Terraform to use the local provider binary

✅ **Resource Creation**
- Creates two `fastly_service_cdn` resources
- Attaches `fastly_service_domain` resources to each service
- Configures multiple `fastly_service_backend` resources
  - Shared backend used by both services
  - Unique backend for service 1

✅ **Data Sources**
- Verifies `data.fastly_service_version` works correctly
- Checks active and latest version numbers

✅ **Resource Updates**
- Tests updating service attributes (comments)
- Validates that changes are applied correctly

✅ **Version Management Actions**
- Tests `fastly_service_version_clone` action
- Tests `fastly_service_version_activate` action
- Validates action invocation workflow

✅ **Resource Destruction**
- Removes version-locked resources from Terraform state
- Uses `terraform destroy` to delete all services
- Verifies services were successfully deleted
- Tests that `force_destroy = true` works correctly
- Emergency cleanup on test failures

### Prerequisites

- **Environment Variables:**
  - `FASTLY_API_TOKEN` - Valid Fastly API token

- **Required Commands:**
  - `terraform` - Terraform CLI
  - `go` - Go toolchain
  - `jq` - JSON processor

### Usage

```bash
# Set your Fastly API token
export FASTLY_API_TOKEN="your_token_here"

# Run the test script from repo root
./scripts/test-orchestration-lifecycle/run.sh

# Or from the test directory
cd scripts/test-orchestration-lifecycle
./run.sh
```

### Output

The script provides colored output showing:
- **Blue** - Section headers and informational messages
- **Green** - Success messages
- **Yellow** - Warnings
- **Red** - Errors

Example output:
```
=== Checking prerequisites ===

[SUCCESS] FASTLY_API_TOKEN is set
[SUCCESS] Found command: terraform
[SUCCESS] Found command: go
[SUCCESS] Found command: jq

=== Building Terraform provider ===

[SUCCESS] Provider built successfully

...

=== Test Summary ===

[SUCCESS] ✓ Provider build
[SUCCESS] ✓ Service creation (fastly_service_cdn)
[SUCCESS] ✓ Domain attachment (fastly_service_domain)
[SUCCESS] ✓ Backend configuration (fastly_service_backend)
[SUCCESS] ✓ Version data sources (data.fastly_service_version)
[SUCCESS] ✓ Resource updates
[SUCCESS] ✓ Version clone action (fastly_service_version_clone)
[SUCCESS] ✓ Version activate action (fastly_service_version_activate)
[SUCCESS] ✓ Resource destruction

[SUCCESS] All orchestration lifecycle tests passed!
```

### Test Artifacts

The script creates a temporary test directory at:
```
/path/to/repo/test-orchestration-<PID>
```

This directory contains:
- Terraform configuration files (main.tf, variables.tf, outputs.tf)
- Terraform state
- Provider configuration

The directory is automatically cleaned up when the script exits.

### Action Testing

Terraform Actions (`fastly_service_version_clone` and `fastly_service_version_activate`) are fully tested by this suite:

1. Actions are invoked explicitly via `terraform apply -invoke=action.X`
2. The script validates that actions execute successfully
3. Version cloning creates new editable service versions
4. Version activation sets the active version

See the detailed README in `test-orchestration-lifecycle/` for more information.

### Failure Handling

If the script fails:
1. Error messages are displayed in red
2. Cleanup still runs automatically
3. Temporary directory is removed
4. Exit code indicates failure

To debug failures, you can comment out the cleanup trap and inspect the test directory manually.

### Integration with CI/CD

This script is ideal for CI/CD pipelines:

```bash
# In CI
export FASTLY_API_TOKEN="${FASTLY_TOKEN_SECRET}"
./scripts/test-orchestration-lifecycle.sh

# Exit code 0 = success
# Exit code 1 = failure
```

### Resources Created

The script creates:
- 2 Fastly CDN services
- 2 domain attachments  
- 3 backend configurations (1 shared, 2 unique)

All resources are tagged with the process ID for uniqueness and are destroyed during cleanup.

### Related Files

- `/examples/orchestration-explicit-actions/` - Full orchestration example
- `/internal/acceptance_tests/service_cdn_test.go` - Unit tests
- `/internal/resources/servicecdn/resource.go` - Resource implementation
