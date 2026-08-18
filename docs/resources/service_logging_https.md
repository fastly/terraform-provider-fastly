---
page_title: "fastly_service_logging_https Resource - fastly"
subcategory: ""
description: |-
  Fastly service HTTPS logging endpoint resource. Writes directly to the specified writable service version.
---

# fastly_service_logging_https (Resource)

Fastly service HTTPS logging endpoint resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages an HTTPS real-time logging endpoint on the configured service version.
It does not clone, activate, or stage service versions.

To have the provider manage the version lifecycle for you instead, use the
nested `logging_https` block on `fastly_service_cdn_auto` or
`fastly_service_compute_auto` — see "Automatic-lifecycle usage" below.

## Example Usage

```terraform
resource "fastly_service_logging_https" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "https-endpoint"
  url        = "https://logs.example.com/ingest"
}
```

A fully configured endpoint using mutual TLS, a custom header, and gzip
compression, with a custom request batching limit. `format`, `format_version`,
`placement`, and `response_condition` only affect generated VCL, so they are
valid when `service_id` refers to a CDN (VCL) service and rejected for a
Compute service:

```terraform
resource "fastly_service_logging_https" "tls" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "https-tls"
  url        = "https://logs.example.com/ingest"

  tls = {
    ca_cert     = file("logs-ca.pem")
    client_cert = file("logs-client.pem")
    client_key  = file("logs-client-key.pem")
    hostname    = "logs.example.com"
  }
  header_name  = "Authorization"
  header_value = var.logs_auth_header
  gzip_level   = 6
  method       = "PUT"

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
resource "fastly_service_logging_https" "compute" {
  service_id = fastly_service_compute.example.id
  version    = 1
  name       = "https-compute"
  url        = "https://logs.example.com/ingest"
}
```

## Automatic-lifecycle usage

Inside the `_auto` service resources, HTTPS logging is a nested block and the
provider clones, validates, and activates a new service version whenever the
block changes. The nested block takes the same arguments as this resource, minus
`service_id` and `version`, which the parent service owns.

```terraform
resource "fastly_service_cdn_auto" "example" {
  name = "my-service"

  domain {
    name = "www.example.com"
  }

  logging_https {
    name = "https-endpoint"
    url  = "https://logs.example.com/ingest"
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

  logging_https {
    name = "https-compute"
    url  = "https://logs.example.com/ingest"
  }
}
```

## Schema

### Required

- `name` (String) The name for the real-time logging configuration. Must be unique within the service.
- `service_id` (String) Fastly service ID.
- `url` (String) URL that log data will be sent to. Must use the HTTPS protocol.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `compression_codec` (String) The codec used for compressing your logs. Valid values are `zstd`, `snappy`, and `gzip`. If the codec is `gzip`, `gzip_level` defaults to `3`; to use a different level, leave `compression_codec` unset and set `gzip_level` instead. Conflicts with `gzip_level`: setting both in the same request will result in an error.
- `content_type` (String) Value of the `Content-Type` header sent with the request.
- `format` (String) A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/).
- `format_version` (Number) The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.
- `gzip_level` (Number) The level of gzip encoding when sending logs. Valid values are `0` (no compression) through `9`. To compress at a specific gzip level, leave `compression_codec` unset and set this. Conflicts with `compression_codec`: setting both in the same request will result in an error.
- `header_name` (String) Custom header sent with the request.
- `header_value` (String) Value of the custom header sent with the request.
- `json_format` (String) Enforces valid JSON formatting for log entries. Can be either disabled (`0`), array of JSON (`1`), or newline delimited JSON (`2`). Default `0`.
- `message_type` (String) How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `blank`.
- `method` (String) HTTP method used for request. Can be either `POST` or `PUT`. Default `POST`.
- `period` (Number) How frequently, in seconds, batches of log data are sent to the HTTPS endpoint. A value of `0` sends logs at the same interval as the default, which is `5` seconds.
- `placement` (String) Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.
- `processing_region` (String) The geographic region where the logs will be processed before streaming. Valid values are `us`, `eu`, and `none` for global. Default: `none`.
- `request_max_bytes` (Number) The maximum number of bytes sent in one request. Default `0` for unbounded (100MB).
- `request_max_entries` (Number) The maximum number of logs sent in one request. Default `0` for unbounded (10k).
- `response_condition` (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- `tls` (Attributes) TLS configuration used to authenticate the HTTPS server, and optionally this endpoint via mutual TLS. (see [below for nested schema](#nestedatt--tls))

### Read-Only

- `id` (String) Terraform resource identifier.

<a id="nestedatt--tls"></a>
### Nested Schema for `tls`

Optional:

- `ca_cert` (String) A secure certificate to authenticate the server with. Must be in PEM format.
- `client_cert` (String) The client certificate used to make authenticated requests. Must be in PEM format.
- `client_key` (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format.
- `hostname` (String) The hostname used to verify the server's certificate. This should be one of the Subject Alternative Name (SAN) fields for the certificate. Common Names (CN) are not supported.

## Import

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the endpoint from the Fastly API and
populate full state:

```shell
terraform import fastly_service_logging_https.example SERVICE_ID/VERSION/ENDPOINT_NAME
```

Example:

```shell
terraform import fastly_service_logging_https.example SU1Z0isxPaozGVKXdv0eY/3/https-endpoint
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.

## Notes

- `tls` groups the certificate material used for mutual TLS to the HTTPS
  server. None of its fields have an environment variable fallback; all default
  to an empty string when omitted. `client_key` is sensitive and never appears
  in plan output.
- `gzip_level` and `compression_codec` are mutually exclusive — the Fastly API
  rejects a request that sets both. Leave `compression_codec` unset and set
  `gzip_level` for a specific level, or set `compression_codec` (`zstd`,
  `snappy`, or `gzip`) and let the API manage the level.
- `header_name`/`header_value` add one custom HTTP header to every request the
  endpoint sends — for example an `Authorization` header some HTTPS log
  collectors require. Neither is treated as sensitive by this resource, since
  the Fastly API accepts arbitrary header values here, not only credentials;
  mark the input variable holding `header_value` sensitive yourself if you use
  it for an authorization token.
- Leaving `placement` unset is not the same as setting it to `none`: unset lets
  Fastly place the logging call automatically (`vcl_log` for `format_version` 2,
  `vcl_deliver` for `format_version` 1), while `none` suppresses the generated
  log statement entirely so you can write it yourself.
