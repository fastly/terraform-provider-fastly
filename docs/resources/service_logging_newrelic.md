---
page_title: "fastly_service_logging_newrelic Resource - fastly"
subcategory: ""
description: |-
  Fastly service New Relic logging endpoint resource. Writes directly to the specified writable service version.
---

# fastly_service_logging_newrelic (Resource)

Fastly service New Relic logging endpoint resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a New Relic real-time logging endpoint on the configured service
version. It does not clone, activate, or stage service versions.

To have the provider manage the version lifecycle for you instead, use the
nested `logging_newrelic` block on `fastly_service_cdn_auto` or
`fastly_service_compute_auto` — see "Automatic-lifecycle usage" below.

## Example Usage

```terraform
resource "fastly_service_logging_newrelic" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "newrelic-us"

  authentication = {
    token = var.newrelic_api_key
  }
  region            = "US"
  processing_region = "us"
}
```

A fully configured endpoint. `format`, `format_version`, `placement`, and
`response_condition` only affect generated VCL, so they are valid when
`service_id` refers to a CDN (VCL) service and rejected for a Compute service:

```terraform
resource "fastly_service_logging_newrelic" "eu" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "newrelic-eu"

  authentication = {
    token = var.newrelic_api_key
  }
  region            = "EU"
  processing_region = "eu"

  format             = "%h %l %u %t \"%r\" %>s %b"
  format_version     = 2
  placement          = "none"
  response_condition = fastly_service_condition.errors_only.name
}
```

Attaching to a Compute service — the VCL-only attributes must be omitted:

```terraform
resource "fastly_service_logging_newrelic" "compute" {
  service_id = fastly_service_compute.example.id
  version    = 1
  name       = "newrelic-compute"

  authentication = {
    token = var.newrelic_api_key
  }
  region = "US"
}
```

## Automatic-lifecycle usage

Inside the `_auto` service resources, New Relic logging is a nested block and
the provider clones, validates, and activates a new service version whenever
the block changes. The nested block takes the same arguments as this resource,
minus `service_id` and `version`, which the parent service owns.

```terraform
resource "fastly_service_cdn_auto" "example" {
  name = "my-service"

  domain {
    name = "www.example.com"
  }

  logging_newrelic {
    name = "newrelic-us"
    authentication = {
      token = var.newrelic_api_key
    }
    region            = "US"
    processing_region = "us"
  }
}
```

`fastly_service_compute_auto` supports the same block, without the VCL-only
arguments (`format`, `format_version`, `placement`, `response_condition`):

```terraform
resource "fastly_service_compute_auto" "example" {
  name = "my-compute-service"

  domain {
    name = "www.example.com"
  }

  package {
    filename         = "package.tar.gz"
    source_code_hash = filesha512("package.tar.gz")
  }

  logging_newrelic {
    name = "newrelic-compute"
    authentication = {
      token = var.newrelic_api_key
    }
    region = "US"
  }
}
```

## Schema

### Required

- `authentication` (Attributes) New Relic authentication credentials. (see [below for nested schema](#nestedatt--authentication))
- `name` (String) The name for the real-time logging configuration. Must be unique within the service.
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `format` (String) A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/). Must produce valid JSON that New Relic Logs can ingest.
- `format_version` (Number) The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.
- `placement` (String) Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.
- `processing_region` (String) Region where logs will be processed before streaming to New Relic. Valid values are `none`, `us` and `eu`.
- `region` (String) The region that log data will be sent to. Valid values are `US` and `EU`. Default: `US`.
- `response_condition` (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### Read-Only

- `id` (String) Terraform resource identifier.

<a id="nestedatt--authentication"></a>
### Nested Schema for `authentication`

Required:

- `token` (String, Sensitive) The Insert API key from the Account page of your New Relic account.

## Import

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the endpoint from the Fastly API and
populate full state:

```shell
terraform import fastly_service_logging_newrelic.example SERVICE_ID/VERSION/ENDPOINT_NAME
```

Example:

```shell
terraform import fastly_service_logging_newrelic.example SU1Z0isxPaozGVKXdv0eY/3/newrelic-us
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.

## Notes

- `authentication` groups credentials as the other logging endpoints do. It is
  required — there is no `FASTLY_NEWRELIC_*` environment variable to default
  from. `token` is sensitive and never appears in plan output.
- Leaving `placement` unset is not the same as setting it to `none`: unset lets
  Fastly place the logging call automatically (`vcl_log` for `format_version` 2,
  `vcl_deliver` for `format_version` 1), while `none` suppresses the generated
  log statement entirely so you can write it yourself.
- `region` accepts only `US` and `EU`. This is the standard New Relic Logs
  endpoint; for the OTLP-based endpoint see `fastly_service_logging_newrelicotlp`.
