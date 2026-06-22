# Terraform Provider Testing Guide

This comprehensive guide covers all aspects of testing the Fastly Terraform Provider.

## Table of Contents

- [Quick Start](#quick-start)
- [Test Organization](#test-organization)
- [Unit Tests](#unit-tests)
- [Acceptance Tests](#acceptance-tests)
- [Template-Based Configuration](#template-based-configuration)
- [Writing New Tests](#writing-new-tests)
- [Debugging Tests](#debugging-tests)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)
- [Reference](#reference)

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

# Acceptance tests (requires API token)
make test-acc

# Run specific resource tests
go test ./internal/resources/backend -v
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend

# Run with coverage
go test ./internal/resources/backend -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Organization

### Directory Structure

```
internal/
├── acceptance_tests/              # Acceptance test framework
│   ├── blocks/                   # Terraform config templates
│   │   ├── backend_basic.tf     # Basic backend template
│   │   ├── backend_updated.tf   # Updated backend template
│   │   ├── backend_full.tf      # Full backend config
│   │   ├── backend_multi.tf     # Multiple backends
│   │   ├── domain_single.tf     # Domain template
│   │   ├── service_cdn.tf       # CDN service template
│   │   └── ...                  # Other templates
│   ├── fixtures/                # Test fixture files
│   │   └── packages/
│   │       └── valid.tar.gz     # Compute package
│   ├── backend_test.go          # Backend acceptance tests
│   ├── backend_import_test.go   # Backend import tests
│   ├── config_builder.go        # Template builder
│   └── test_helpers.go          # Shared infrastructure
└── resources/
    └── backend/
        ├── backend_test.go      # Unit tests
        ├── resource.go         # Implementation
        ├── schema.go           # Schema definition
        ├── expand.go           # TF → API conversion
        └── flatten.go          # API → TF conversion
```

### Test Infrastructure Components

**Core Infrastructure** (`test_helpers.go`):
- `ProtoV6ProviderFactories()` - Provider factories for tests
- `PreCheck(t)` - Environment validation
- `NewFastlyClient()` - Fastly API client creation
- `CheckServiceDestroy()` - Verify service deletion
- `CheckServiceExists()` - Verify service existence
- `CheckBackendExistsInFastly()` - Verify backend in API

**Template Builder** (`config_builder.go`):
- `BuildConfig()` - Compose Terraform configs from templates
- Uses Go `text/template` for safe placeholder replacement
- Templates in `blocks/` directory with `{{.PLACEHOLDER}}` format

## Unit Tests

Unit tests validate individual functions without making API calls. They should achieve 100% coverage of expand, flatten, and schema functions.

### What to Test

1. **expand.go functions**:
   - `BuildCreateInput()` - All fields, minimal config, edge cases
   - `BuildUpdateInput()` - Field changes, defaults, nulls

2. **flatten.go functions**:
   - `flatten()` - Service metadata, ID generation
   - `FlattenToNestedModel()` - All fields, nil values, empty values

3. **schema.go functions**:
   - `ModelsEqual()` - Field-by-field comparison
   - `Equal()` - Slice comparison, order independence
   - `Normalize()` - Sorting, stability
   - `Reconcile()` - (if applicable)

### Unit Test Pattern

```go
func TestBuildCreateInput(t *testing.T) {
    tests := []struct {
        name      string
        serviceID string
        version   int
        model     NestedModel
        validate  func(t *testing.T, input *fastly.CreateBackendInput)
    }{
        {
            name:      "minimal backend",
            serviceID: "service-123",
            version:   1,
            model: NestedModel{
                Name:    types.StringValue("origin"),
                Address: types.StringValue("api.example.com"),
            },
            validate: func(t *testing.T, input *fastly.CreateBackendInput) {
                assert.Equal(t, "service-123", input.ServiceID)
                assert.Equal(t, 1, input.ServiceVersion)
                assert.Equal(t, "origin", *input.Name)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := BuildCreateInput(tt.serviceID, tt.version, tt.model)
            tt.validate(t, input)
        })
    }
}
```

### Running Unit Tests

```bash
# All unit tests for a resource
go test ./internal/resources/backend -v

# Specific test categories
go test ./internal/resources/backend -v -run TestBuild
go test ./internal/resources/backend -v -run TestFlatten
go test ./internal/resources/backend -v -run "TestEqual|TestModels|TestNormalize"

# Individual test
go test ./internal/resources/backend -v -run TestModelsEqual

# With coverage
go test ./internal/resources/backend -v -coverprofile=coverage.out
go tool cover -html=coverage.out

# With race detection
go test ./internal/resources/backend -v -race
```

## Acceptance Tests

Acceptance tests validate end-to-end functionality against the live Fastly API.

⚠️ **Warning**: These tests create real resources in your Fastly account and consume API rate limits. Resources are cleaned up automatically with `force_destroy = true`.

### Prerequisites

1. **Fastly API Token**: 
   ```bash
   export FASTLY_API_TOKEN="your-token-here"
   ```

2. **Enable Acceptance Tests**:
   ```bash
   export TF_ACC="1"
   ```

3. **Account Permissions**: Token must have:
   - Create/read/update/delete services and versions
   - Activate service versions
   - Create/read/update/delete all resource types being tested

### What to Test

1. **Basic CRUD**: Create, read, update, delete with minimal config
2. **Full Configuration**: All optional fields populated
3. **Multiple Resources**: Multiple instances don't interfere
4. **Update Scenarios**: Change required/optional fields
5. **Edge Cases**: Mutability checks, service types, special characters
6. **Import**: Both legacy composite ID and identity schema

### Acceptance Test Pattern

```go
func TestAccFastlyServiceBackend_basic(t *testing.T) {
    t.Parallel()
    serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
    domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
    backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { PreCheck(t) },
        ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
        CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
        Steps: []resource.TestStep{
            {
                Config: ConfigBackendBasic(serviceName, domainName, backendName),
                Check: resource.ComposeTestCheckFunc(
                    CheckServiceExists("fastly_service_cdn.test"),
                    resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
                    resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.example.com"),
                    CheckBackendExistsInFastly("fastly_service_cdn.test", backendName, 1),
                ),
            },
        },
    })
}
```

### Running Acceptance Tests

```bash
# All acceptance tests
make test-acc

# Specific resource tests
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend

# Specific test
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend_basic

# With parallelism
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend -parallel 4 -timeout 30m
```

### Best Practices

1. **Use `t.Parallel()`** for all tests (except those modifying env vars)
2. **Generate unique names** with `acctest.RandString()`
3. **Clean up resources** with `force_destroy = true`
4. **Test edge cases** including empty strings, null values, special characters
5. **Verify API state** beyond Terraform state using check functions
6. **Test mutability** by activating versions

## Template-Based Configuration

All test configurations use the `BuildConfig()` template system for maintainability and reusability.

### How BuildConfig Works

1. Loads service template from `blocks/{serviceType}.tf`
2. Parses nested block templates using `text/template`
3. Injects rendered blocks into `{{.RESOURCES}}` placeholder
4. Replaces all placeholders with values from replacements map

### Using BuildConfig

```go
func ConfigBackendBasic(serviceName, domainName, backendName string) string {
    return BuildConfig(
        ServiceCDN,
        map[string]string{
            "SERVICE_NAME":    serviceName,
            "SERVICE_COMMENT": "",
            "DOMAIN_NAME":     domainName,
            "SERVICE_VERSION": "1",
            "BACKEND_NAME":    backendName,
        },
        "internal/acceptance_tests/blocks/service_cdn_domain.tf",
        "internal/acceptance_tests/blocks/backend_basic.tf",
    )
}
```

### Available Configuration Helpers

#### Auto Services
- `ConfigCDNAutoBasic(serviceName, domainName)`
- `ConfigCDNAutoWithBackend(serviceName, domainName, backendName)`
- `ConfigCDNAutoMultipleBackends(serviceName, domainName)`
- `ConfigComputeAutoBasic(serviceName, domainName)`
- `ConfigComputeAutoWithBackend(serviceName, domainName, backendName)`
- `ConfigComputeAutoMultipleBackends(serviceName, domainName)`

#### Explicit Services
- `ConfigServiceCDNBasic(serviceName)`
- `ConfigServiceCDNWithDomain(serviceName, domainName, version)`
- `ConfigServiceCDNWithBackend(serviceName, domainName, backendName, version)`
- `ConfigServiceCDNWithVersionClone(serviceName, domainName)`
- `ConfigServiceCDNWithVersionActivate(serviceName, domainName, version)`

#### Backend Resources
- `ConfigBackendBasic(serviceName, domainName, backendName)`
- `ConfigBackendUpdated(serviceName, domainName, backendName)`
- `ConfigBackendFull(serviceName, domainName, backendName)`
- `ConfigBackendMultiple(serviceName, domainName, backend1, backend2)`
- `ConfigBackendForImport(serviceName, domainName, backendName)`

### Available Templates

**Service Templates**:
- `service_cdn.tf` - Explicit CDN (manual version management)
- `service_cdn_auto.tf` - Automatic CDN
- `service_compute.tf` - Explicit Compute
- `service_compute_auto.tf` - Automatic Compute

**Resource Templates**:
- `backend_basic.tf` - Minimal backend
- `backend_updated.tf` - Updated backend
- `backend_full.tf` - All optional fields
- `backend_multi.tf` - Multiple backends
- `domain_single.tf` - Single domain
- `service_cdn_domain.tf` - Domain resource (explicit)
- `service_cdn_backend.tf` - Backend resource (explicit)

**Action Templates**:
- `action_version_clone.tf` - Version clone
- `action_version_activate.tf` - Version activate

## Writing New Tests

### Adding Tests for a New Resource

1. **Create template blocks** in `internal/acceptance_tests/blocks/`:
   ```
   <resource>_basic.tf
   <resource>_updated.tf (optional)
   <resource>_full.tf (optional)
   ```

2. **Add config helpers** to `test_helpers.go`:
   ```go
   func Config<Resource>Basic(...) string {
       return BuildConfig(...)
   }
   ```

3. **Create unit tests** in `internal/resources/<name>/<name>_test.go`:
   - Test all expand.go functions
   - Test all flatten.go functions
   - Test all schema.go functions

4. **Create acceptance tests** in `internal/acceptance_tests/<name>_test.go`:
   - TestAcc<Resource>_basic
   - TestAcc<Resource>_update
   - TestAcc<Resource>_fullConfig
   - TestAcc<Resource>_multiple (if applicable)

5. **Create import tests** in `internal/acceptance_tests/<name>_import_test.go`:
   - TestAcc<Resource>_importBasic
   - TestAcc<Resource>_importWithEdgeCases

### Test Naming Conventions

- Unit tests: `Test<FunctionName>`
- Acceptance tests: `TestAcc<ResourceType>_<scenario>`
- Use snake_case for scenario names
- Be descriptive: `TestAccFastlyServiceBackend_updateChangesAddress`

## Debugging Tests

### Enable Terraform Logs

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend_basic
cat terraform.log
```

### Run Single Test

```bash
TF_ACC=1 go test ./internal/acceptance_tests -v -run TestAccFastlyServiceBackend_basic
```

### Inspect API State

```go
PreConfig: func() {
    client, _ := NewFastlyClient()
    backend, _ := client.GetBackend(ctx, &fastly.GetBackendInput{
        ServiceID:      serviceID,
        ServiceVersion: 1,
        Name:           backendName,
    })
    fmt.Printf("API State: %+v\n", backend)
}
```

## Troubleshooting

### "FASTLY_API_TOKEN must be set"

Verify token is exported:
```bash
echo $FASTLY_API_TOKEN
```

If empty, export it:
```bash
export FASTLY_API_TOKEN="your-token"
```

Verify with Fastly CLI:
```bash
fastly whoami
```

### Tests Timeout

Increase timeout:
```bash
TF_ACC=1 go test ./internal/acceptance_tests -v -timeout 60m
```

### Rate Limiting

Reduce parallel execution:
```bash
TF_ACC=1 go test ./internal/acceptance_tests -v -parallel 2
```

Or add delays between test runs.

### Stale Test Resources

Clean up manually:
```bash
# List services with tf-test prefix
fastly service list | grep tf-test

# Delete specific service
fastly service delete --service-id=<id> --force
```

Or via Fastly dashboard:
1. Go to https://manage.fastly.com/services/
2. Search for "tf-test-"
3. Delete stale services

### Permission Issues

Ensure API token has permissions to:
- Create/read/update/delete services and versions
- Activate versions
- Create/read/update/delete all resource types being tested

## Reference

### Test Coverage Goals

- **Unit tests**: 80%+ coverage for expand, flatten, schema functions
- **Acceptance tests**: 100% coverage of CRUD operations
- **Import tests**: Both legacy and identity schema formats
- **Edge cases**: Special characters, null values, empty strings

### Example Test Coverage (Backend)

**Unit Tests** (`internal/resources/backend/backend_test.go`):
- 9 test functions
- 46 test cases
- 100% function coverage for expand/flatten
- All schema comparison functions tested

**Acceptance Tests** (`internal/acceptance_tests/backend_test.go`):
- 6 CRUD tests
- 2 import tests
- All CRUD operations covered
- Edge cases: mutability, service types, special characters

### Lifecycle Tests

End-to-end lifecycle tests in `scripts/test-lifecycle-{cdn,compute}/`:

```bash
# CDN service lifecycle test
make test-lifecycle-cdn

# Compute service lifecycle test
make test-lifecycle-compute

# All lifecycle tests
make test-lifecycle
```

These test full workflows including version cloning, activation, and resource management.

### External Resources

- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Fastly Go Client](https://github.com/fastly/go-fastly)
- [Compute@Edge Documentation](https://www.fastly.com/documentation/guides/compute/)
