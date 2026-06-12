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
├── blocks/                          # Terraform config blocks
│   ├── backend_single.tf
│   ├── backend_multiple.tf
│   ├── domain_single.tf
│   ├── package.tf
│   ├── service_cdn_auto.tf
│   └── service_compute_auto.tf
├── fixtures/                        # Test fixture files
│   └── packages/
│       └── valid.tar.gz            # WebAssembly package for compute tests
├── config_builder.go                # Config template builder
├── test_helpers.go                  # Shared test helpers
```

### Shared Test Helpers

Common functionality is centralized in `test_helpers.go`:

- **`ProtoV6ProviderFactories()`** - Provider factories for tests
- **`PreCheck(t)`** - Validates FASTLY_API_TOKEN is set
- **`NewFastlyClient()`** - Creates Fastly API client
- **`CheckServiceDestroy(resourceType)`** - Verifies service destruction
- **`CheckServiceExists(resourceName)`** - Verifies service exists
- **`GetPackagePath()`** - Returns path to test package

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

Test configurations are built dynamically using `BuildConfig()` within the `acceptancetests` package:

```go
func configYourServiceBasic(serviceName, domainName string) string {
    return BuildConfig(
        ServiceCDNAuto,  // or ServiceComputeAuto
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


### Available Service Templates

- **`service_cdn_auto.tf`** - CDN service resource
- **`service_compute_auto.tf`** - Compute service resource

### Available Configuration Blocks

- **`domain_single.tf`** - Single domain configuration
- **`backend_single.tf`** - Single backend configuration
- **`backend_multiple.tf`** - Multiple backend configuration
- **`package.tf`** - Compute package configuration

The `BuildConfig` function:
1. Loads the service template
2. Injects the specified blocks into the `RESOURCES` placeholder
3. Replaces all placeholders (`SERVICE_NAME`, `DOMAIN_NAME`, etc.) with actual values

## References

- [Terraform Plugin Testing Documentation](https://developer.hashicorp.com/terraform/plugin/testing)
- [Fastly Go Client Documentation](https://github.com/fastly/go-fastly)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Fastly Compute Starter Kits](https://developer.fastly.com/solutions/starters/)
- [Compute@Edge Documentation](https://www.fastly.com/documentation/guides/compute/)
