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

### Test Coverage

**Backend Resource** (`fastly_service_backend`)
- Basic configuration with defaults
- Field updates (address, port, weight, comment)
- Complete SSL/TLS, timeouts, load balancing configuration
- Import functionality
- SSL certificate fields

**CDN Auto Resource** (`fastly_service_cdn_auto`)
- Basic service with domain
- Service with backend configuration
- Updates (name, comment, multiple domains)
- Multiple backends with load balancing
- Import functionality

**Compute Auto Resource** (`fastly_service_compute_auto`)
- Basic service with domain and package
- Service with backend configuration
- Updates (name, comment, multiple domains)
- Multiple backends with load balancing
- Import functionality

### Cleanup

If tests are interrupted and leave resources behind:

```bash
# List services created by tests
fastly service list | grep "tf-test-"

# Delete a specific service
fastly service delete --service-id=<service-id> --force
```

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

```go
func TestAccFastlyServiceBackend_yourTest(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        CheckDestroy:             testAccCheckFastlyServiceBackendDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccFastlyServiceBackendConfig_yourConfig(...),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckFastlyServiceBackendExists("fastly_service_backend.test"),
                    resource.TestCheckResourceAttr("fastly_service_backend.test", "attribute", "expected_value"),
                ),
            },
        },
    })
}
```

## CI/CD Integration

To run these tests in CI/CD:

```yaml
# Example GitHub Actions
- name: Run Unit Tests
  run: make test-unit

- name: Run Acceptance Tests
  env:
    FASTLY_API_TOKEN: ${{ secrets.FASTLY_API_TOKEN }}
  run: make test-acc
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

## Test Configuration Templates

Test configurations are stored as templates in `test_fixtures/configs/`:

- **`basic.tf`** - Minimal service with domain and package
- **`with_backend.tf`** - Service with backend configuration  
- **`updated.tf`** - Multiple domains with comment
- **`multiple_backends.tf`** - Service with multiple backends

Templates use placeholder strings (`SERVICE_NAME`, `DOMAIN_NAME`, `BACKEND_NAME`) that are replaced at test runtime using `strings.ReplaceAll()`:

```go
func testAccServiceComputeAutoConfig_basic(serviceName, domainName string) string {
    return strings.ReplaceAll(strings.ReplaceAll(basicTemplate, 
        "SERVICE_NAME", serviceName), 
        "DOMAIN_NAME", domainName)
}
```

## References

- [Terraform Plugin Testing Documentation](https://developer.hashicorp.com/terraform/plugin/testing)
- [Fastly Go Client Documentation](https://github.com/fastly/go-fastly)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Fastly Compute Starter Kits](https://developer.fastly.com/solutions/starters/)
- [Compute@Edge Documentation](https://www.fastly.com/documentation/guides/compute/)
