---
page_title: "fastly_service_logging_splunk Resource - fastly"
subcategory: ""
description: |-
  Fastly service Splunk logging endpoint resource. Writes directly to the specified writable service version.
---

# fastly_service_logging_splunk (Resource)

Fastly service Splunk logging endpoint resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a Splunk real-time logging endpoint on the configured service version.
It does not clone, activate, or stage service versions.

To have the provider manage the version lifecycle for you instead, use the
nested `logging_splunk` block on `fastly_service_cdn_auto` or
`fastly_service_compute_auto` — see "Automatic-lifecycle usage" below.

## Example Usage

```terraform
resource "fastly_service_logging_splunk" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "splunk-endpoint"
  url        = "https://splunk.example.com/services/collector/event"

  authentication = {
    token = var.splunk_token
  }
}
```

A fully configured endpoint using mutual TLS, with a custom request batching
limit. `format`, `format_version`, `placement`, and `response_condition` only
affect generated VCL, so they are valid when `service_id` refers to a CDN (VCL)
service and rejected for a Compute service:

```terraform
resource "fastly_service_logging_splunk" "tls" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "splunk-tls"
  url        = "https://splunk.example.com/services/collector/event"

  authentication = {
    token = var.splunk_token
  }
  tls = {
    ca_cert     = file("splunk-ca.pem")
    client_cert = file("splunk-client.pem")
    client_key  = file("splunk-client-key.pem")
    hostname    = "splunk.example.com"
  }
  use_tls             = true
  request_max_bytes   = 1000000
  request_max_entries = 1000

  format             = "%h %l %u %t \"%r\" %>s %b"
  format_version     = 2
  placement          = "none"
  response_condition = fastly_service_condition.errors_only.name
}
```

Attaching to a Compute service — the VCL-only attributes must be omitted:

```terraform
resource "fastly_service_logging_splunk" "compute" {
  service_id = fastly_service_compute.example.id
  version    = 1
  name       = "splunk-compute"
  url        = "https://splunk.example.com/services/collector/event"

  authentication = {
    token = var.splunk_token
  }
}
```

## Automatic-lifecycle usage

Inside the `_auto` service resources, Splunk logging is a nested block and the
provider clones, validates, and activates a new service version whenever the
block changes. The nested block takes the same arguments as this resource, minus
`service_id` and `version`, which the parent service owns.

```terraform
resource "fastly_service_cdn_auto" "example" {
  name = "my-service"

  domain {
    name = "www.example.com"
  }

  logging_splunk {
    name = "splunk-endpoint"
    url  = "https://splunk.example.com/services/collector/event"
    authentication = {
      token = var.splunk_token
    }
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

  logging_splunk {
    name = "splunk-compute"
    url  = "https://splunk.example.com/services/collector/event"
    authentication = {
      token = var.splunk_token
    }
  }
}
```

## Schema

### Required

- `name` (String) The name for the real-time logging configuration. Must be unique within the service.
- `service_id` (String) Fastly service ID.
- `url` (String) The URL to post logs to.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `authentication` (Attributes) Splunk authentication credentials. When this block is omitted entirely, defaults to the `FASTLY_SPLUNK_TOKEN` environment variable. (see [below for nested schema](#nestedatt--authentication))
- `format` (String) A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/).
- `format_version` (Number) The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.
- `placement` (String) Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.
- `processing_region` (String) The geographic region where the logs will be processed before streaming. Valid values are `us`, `eu`, and `none` for global. Default: `none`.
- `request_max_bytes` (Number) The maximum number of bytes sent in one request. Default `0` for unbounded.
- `request_max_entries` (Number) The maximum number of logs sent in one request. Default `0` for unbounded.
- `response_condition` (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- `tls` (Attributes) TLS configuration used when `use_tls` is enabled. When this block is omitted entirely, `ca_cert`, `client_cert`, and `client_key` default to the `FASTLY_SPLUNK_CA_CERT`, `FASTLY_SPLUNK_CLIENT_CERT`, and `FASTLY_SPLUNK_CLIENT_KEY` environment variables. (see [below for nested schema](#nestedatt--tls))
- `use_tls` (Boolean) Whether to use TLS for secure logging. Default: `false`.

### Read-Only

- `id` (String) Terraform resource identifier.

<a id="nestedatt--authentication"></a>
### Nested Schema for `authentication`

Optional:

- `token` (String, Sensitive) A Splunk token for use in posting logs over HTTP to your collector. Can be set via the `FASTLY_SPLUNK_TOKEN` environment variable.


<a id="nestedatt--tls"></a>
### Nested Schema for `tls`

Optional:

- `ca_cert` (String) A secure certificate to authenticate the server with. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CA_CERT` environment variable.
- `client_cert` (String) The client certificate used to make authenticated requests. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CLIENT_CERT` environment variable.
- `client_key` (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CLIENT_KEY` environment variable.
- `hostname` (String) The hostname used to verify the server's certificate. This should be one of the Subject Alternative Name (SAN) fields for the certificate. Common Names (CN) are not supported.

## Import

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the endpoint from the Fastly API and
populate full state:

```shell
terraform import fastly_service_logging_splunk.example SERVICE_ID/VERSION/ENDPOINT_NAME
```

Example:

```shell
terraform import fastly_service_logging_splunk.example SU1Z0isxPaozGVKXdv0eY/3/splunk-endpoint
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.

## Notes

- `authentication` groups the Splunk token as the other logging endpoints group
  credentials. Unlike some of the other logging endpoints, it is Optional and
  Computed: when omitted entirely it defaults to the `FASTLY_SPLUNK_TOKEN`
  environment variable, preserving the live (SDKv2) provider's behavior.
  `token` is sensitive and never appears in plan output.
- `tls` groups the certificate material used for mutual TLS to the Splunk
  collector. `ca_cert` and `client_cert` default to the `FASTLY_SPLUNK_CA_CERT`
  and `FASTLY_SPLUNK_CLIENT_CERT` environment variables when omitted; `client_key`
  defaults to `FASTLY_SPLUNK_CLIENT_KEY` and is sensitive. `hostname` has no
  environment variable and defaults to an empty string.
- Leaving `placement` unset is not the same as setting it to `none`: unset lets
  Fastly place the logging call automatically (`vcl_log` for `format_version` 2,
  `vcl_deliver` for `format_version` 1), while `none` suppresses the generated
  log statement entirely so you can write it yourself.
