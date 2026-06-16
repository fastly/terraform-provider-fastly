# Terraform Provider Testing

This document covers testing for the Fastly Terraform Provider.

## Quick Start

### Setup
```bash
# Set your Fastly API token (required for acceptance tests)
export FASTLY_API_TOKEN="your-token-here"
```

### Running Tests

```bash
# Unit tests (no API token needed)
make test-unit

# Acceptance tests (requires API token above)
make test-acc
```

## Unit Tests

Unit tests validate individual functions in isolation without making API calls.

## Acceptance Tests

Acceptance tests validate end-to-end functionality against the live Fastly API.

⚠️ **Warning**: These tests create real resources in your Fastly account and consume API rate limits. Resources are automatically cleaned up with `force_destroy = true`.

### Cleanup

If tests are interrupted and leave resources behind:

```bash
# List services created by tests using the CLI
fastly service list | grep "tf-test-"

# Delete a specific service
fastly service delete --service-id=<service-id> --force
```

## Test Organization

### Directory Structure

All acceptance tests are consolidated in `internal/acceptance_tests/`:

```
internal/acceptance_tests/
├── blocks/                          # Terraform config templates
│   ├── action_version_activate.tf  # Version activation action
│   ├── action_version_clone.tf     # Version clone action
│   ├── backend_multiple.tf         # Multiple backend blocks
│   ├── backend_single.tf           # Single backend block
│   ├── domain_single.tf            # Single domain block
│   ├── package.tf                  # Compute package block
│   ├── service_cdn.tf              # Explicit CDN service template
│   ├── service_cdn_auto.tf         # Auto CDN service template
│   ├── service_cdn_backend.tf      # Backend resource for explicit CDN
│   ├── service_cdn_domain.tf       # Domain resource for explicit CDN
│   ├── service_compute.tf          # Explicit Compute service template
│   └── service_compute_auto.tf     # Auto Compute service template
├── fixtures/                        # Test fixture files
│   └── packages/
│       └── valid.tar.gz            # WebAssembly package for compute tests
├── config_builder.go                # Config template builder
└── test_helpers.go                  # Shared test helpers
```

### Shared Test Helpers

Common functionality is centralized in `test_helpers.go`:

#### Core Test Infrastructure
- **`ProtoV6ProviderFactories()`** - Provider factories for tests
- **`PreCheck(t)`** - Validates FASTLY_API_TOKEN is set
- **`NewFastlyClient()`** - Creates Fastly API client
- **`GetPackagePath()`** - Returns path to test package

#### Check Functions
- **`CheckServiceDestroy(resourceType)`** - Verifies service destruction
- **`CheckServiceExists(resourceName)`** - Verifies service exists

#### Configuration Helpers - Auto Services
- **`ConfigCDNAutoBasic(serviceName, domainName)`** - CDN auto service with domain
- **`ConfigCDNAutoWithBackend(serviceName, domainName, backendName)`** - CDN auto with backend
- **`ConfigCDNAutoMultipleBackends(serviceName, domainName)`** - CDN auto with multiple backends
- **`ConfigComputeAutoBasic(serviceName, domainName)`** - Compute auto service with domain and package
- **`ConfigComputeAutoWithBackend(serviceName, domainName, backendName)`** - Compute auto with backend
- **`ConfigComputeAutoMultipleBackends(serviceName, domainName)`** - Compute auto with multiple backends

#### Configuration Helpers - Explicit Services
- **`ConfigServiceCDNBasic(serviceName)`** - Basic CDN service (no nested resources)
- **`ConfigServiceCDNWithComment(serviceName)`** - CDN service with comment
- **`ConfigServiceCDNWithDomain(serviceName, domainName, version)`** - CDN service with domain resource
- **`ConfigServiceCDNWithBackend(serviceName, domainName, backendName, version)`** - CDN service with backend
- **`ConfigServiceCDNWithVersionClone(serviceName, domainName)`** - CDN service with clone action
- **`ConfigServiceCDNWithVersionActivate(serviceName, domainName, version)`** - CDN service with activate action
- **`ConfigServiceCDNWithCloneAndActivate(serviceName, domainName, backendName)`** - CDN service with both actions
- **`ConfigServiceComputeBasic(serviceName)`** - Basic Compute service
- **`ConfigServiceComputeWithComment(serviceName)`** - Compute service with comment

## Writing New Tests

### Unit Test Pattern

```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        {
            name: "descriptive test case name",
            input: ...,
            expected: ...,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Acceptance Test Pattern

Acceptance tests should be added to `internal/acceptance_tests/` (package `acceptancetests`) and use the shared helpers:

```go
package acceptancetests

func TestAccFastlyServiceYourResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { PreCheck(t) },
        ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
        CheckDestroy:             CheckServiceDestroy("fastly_service_your_resource"),
        Steps: []resource.TestStep{
            {
                Config: configYourResourceBasic(serviceName, domainName),
                Check: resource.ComposeTestCheckFunc(
                    CheckServiceExists("fastly_service_your_resource.test"),
                    resource.TestCheckResourceAttr("fastly_service_your_resource.test", "name", serviceName),
                ),
            },
        },
    })
}
```

## Test Configuration Builder

Test configurations are built dynamically using `BuildConfig()` within the `acceptancetests` package.

The builder uses Go's `text/template` for safe placeholder replacement. All template files (`.tf` files in the `blocks/` directory) use the `{{.PLACEHOLDER_NAME}}` format.

Example:

```go
func configYourServiceBasic(serviceName, domainName string) string {
    return BuildConfig(
        ServiceCDNAuto,  // or ServiceComputeAuto, ServiceCDN, ServiceCompute
        map[string]string{
            "SERVICE_NAME": serviceName,
            "DOMAIN_NAME":  domainName,
        },
        "internal/acceptance_tests/blocks/domain_single.tf",
        "internal/acceptance_tests/blocks/backend_single.tf",  // optional additional blocks
    )
}
```

## Troubleshooting

### "FASTLY_API_TOKEN must be set"
Verify the token is exported: `fastly whoami`

### Rate limit exceeded
Add delays between test runs or run serially instead of in parallel.

### Resources not cleaned up
Manually delete via Fastly dashboard/CLI or verify `force_destroy = true` is set.

### Tests timeout
Check Fastly API status and network connectivity.

## Lifecycle Tests

End-to-end lifecycle tests are located in `scripts/test-lifecycle-cdn/` and `scripts/test-lifecycle-compute/`. These test the full provider workflow including version cloning, activation, and resource management.

### Running Lifecycle Tests

```bash
# CDN service lifecycle test
./scripts/test-lifecycle-cdn/run.sh

# Compute service lifecycle test
./scripts/test-lifecycle-compute/run.sh
```

Both scripts:
- Build the provider from source
- Create test services with domains and backends
- Test version clone and activate actions
- Test version-locked resource writes
- Clean up all resources after completion

See the README in each script directory for details.

## Test Templates Reference

### Available Service Templates

- **`service_cdn.tf`** - Explicit CDN service resource (manual version management)
- **`service_cdn_auto.tf`** - Automatic CDN service resource
- **`service_compute.tf`** - Explicit Compute service resource (manual version management)
- **`service_compute_auto.tf`** - Automatic Compute service resource

### Available Configuration Blocks

Blocks can be combined with service templates via `BuildConfig()`:

- **`domain_single.tf`** - Single domain configuration
- **`backend_single.tf`** - Single backend configuration
- **`backend_multiple.tf`** - Multiple backend configuration (3 backends)
- **`package.tf`** - Compute package configuration
- **`action_version_clone.tf`** - Version clone action
- **`action_version_activate.tf`** - Version activate action
- **`service_cdn_domain.tf`** - Domain resource for explicit CDN services
- **`service_cdn_backend.tf`** - Backend resource for explicit CDN services

### How BuildConfig Works

The `BuildConfig` function:
1. Loads the service template from `internal/acceptance_tests/blocks/{serviceType}.tf`
2. Parses and renders nested block templates using `text/template`
3. Injects the rendered blocks into the `{{.RESOURCES}}` placeholder
4. Replaces all placeholders (e.g., `{{.SERVICE_NAME}}`, `{{.DOMAIN_NAME}}`) with actual values from the replacements map

All templates use the `{{.PLACEHOLDER_NAME}}` format for safe, explicit substitution.

## References

- [Terraform Plugin Testing Documentation](https://developer.hashicorp.com/terraform/plugin/testing)
- [Fastly Go Client Documentation](https://github.com/fastly/go-fastly)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Fastly Compute Starter Kits](https://developer.fastly.com/solutions/starters/)
- [Compute@Edge Documentation](https://www.fastly.com/documentation/guides/compute/)
