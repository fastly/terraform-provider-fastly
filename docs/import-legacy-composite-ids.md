# Legacy Composite ID Import Support

## Overview

The `fastly_service_backend` and `fastly_service_domain` resources support legacy composite ID imports for backwards compatibility. This allows users to import resources using the format `service_id/version/name`, even though the stable Framework identity only uses `service_id + name`.

## Why This Matters

Backend and domain resources are service-versioned API objects. While the stable identity excludes `version` (so changing versions doesn't make Terraform treat the resource as a different object), the import process still needs the version to perform the initial read from the API.

## Import Formats

### Backend Import

```bash
# Legacy composite ID format (supported for backwards compatibility)
terraform import fastly_service_backend.origin service123/3/origin

# Identity-based import (new format)
terraform import fastly_service_backend.origin '{"service_id":"service123","name":"origin"}'
```

### Domain Import

```bash
# Legacy composite ID format (supported for backwards compatibility)
terraform import fastly_service_domain.www service123/3/www.example.com

# Identity-based import (new format)
terraform import fastly_service_domain.www '{"service_id":"service123","name":"www.example.com"}'
```

## How It Works

1. When a composite ID containing slashes is detected, the provider parses it as `service_id/version/name`
2. The provider uses the version to perform the initial API read
3. The full resource state is populated from the API response
4. The stable identity (service_id + name only) is set, excluding version
5. Future Terraform operations use the stable identity, so version changes don't cause resource replacement

## Edge Cases Handled

- **Names with slashes**: The parser uses `SplitN(..., 3)` so only the first two slashes are delimiters. A backend named `backend/with/slashes` can be imported as `service123/3/backend/with/slashes`.
- **Complex domain names**: Multi-level subdomains like `api.v2.staging.example.com` are preserved correctly.
- **Version 0**: Although unlikely in practice, version 0 is valid and handled correctly.

## Testing

Unit tests: `internal/importutil/composite_id_test.go`
Acceptance tests:
- `internal/acceptance_tests/backend_import_test.go`
- `internal/acceptance_tests/domain_import_test.go`

Run acceptance tests with:
```bash
export FASTLY_API_TOKEN=<token>
go test ./internal/acceptance_tests -run TestAccFastlyServiceBackend_import
go test ./internal/acceptance_tests -run TestAccFastlyServiceDomain_import
```
