---
page_title: "fastly_service_logging_gcs Resource - fastly"
subcategory: ""
description: |-
  Fastly service GCS logging endpoint resource. Writes directly to the specified writable service version.
---

# fastly_service_logging_gcs (Resource)

Fastly service GCS logging endpoint resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a Google Cloud Storage real-time logging endpoint on the configured
service version. It does not clone, activate, or stage service versions.

To have the provider manage the version lifecycle for you instead, use the
nested `logging_gcs` block on `fastly_service_cdn_auto` or
`fastly_service_compute_auto` — see "Automatic-lifecycle usage" below.

## Example Usage

```terraform
resource "fastly_service_logging_gcs" "example" {
  service_id  = fastly_service_cdn.example.id
  version     = 1
  name        = "gcs-example"
  bucket_name = "my-log-bucket"

  authentication = {
    email      = var.gcs_service_account_email
    secret_key = var.gcs_service_account_secret_key
  }

  format = "{\n \"timestamp\":\"%{begin:%Y-%m-%dT%H:%M:%S}t\",\n  \"client_ip\":\"%{req.http.Fastly-Client-IP}V\"\n}"
}
```

A fully configured endpoint, using `account_name` instead of `email`/`secret_key`
to reference a GCP service account already linked to the Fastly account.
`format`, `format_version`, `placement`, and `response_condition` only affect
generated VCL, so they are valid when `service_id` refers to a CDN (VCL)
service and rejected for a Compute service:

```terraform
resource "fastly_service_logging_gcs" "linked_account" {
  service_id  = fastly_service_cdn.example.id
  version     = 1
  name        = "gcs-linked-account"
  bucket_name = "my-log-bucket"
  path        = "/logs/"
  period      = 1800

  authentication = {
    account_name = "my-linked-service-account"
  }
  processing_region = "eu"
  compression_codec = "gzip"
  message_type      = "loggly"

  format             = "{\n \"timestamp\":\"%{begin:%Y-%m-%dT%H:%M:%S}t\"\n}"
  format_version     = 2
  placement          = "none"
  response_condition = fastly_service_condition.errors_only.name
}
```

Attaching to a Compute service — the VCL-only attributes must be omitted:

```terraform
resource "fastly_service_logging_gcs" "compute" {
  service_id  = fastly_service_compute.example.id
  version     = 1
  name        = "gcs-compute"
  bucket_name = "my-log-bucket"

  authentication = {
    email      = var.gcs_service_account_email
    secret_key = var.gcs_service_account_secret_key
  }
}
```

## Automatic-lifecycle usage

Inside the `_auto` service resources, GCS logging is a nested block and the
provider clones, validates, and activates a new service version whenever the
block changes. The nested block takes the same arguments as this resource,
minus `service_id` and `version`, which the parent service owns.

```terraform
resource "fastly_service_cdn_auto" "example" {
  name = "my-service"

  domain {
    name = "www.example.com"
  }

  logging_gcs {
    name        = "gcs-example"
    bucket_name = "my-log-bucket"

    authentication = {
      email      = var.gcs_service_account_email
      secret_key = var.gcs_service_account_secret_key
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

  logging_gcs {
    name        = "gcs-compute"
    bucket_name = "my-log-bucket"

    authentication = {
      email      = var.gcs_service_account_email
      secret_key = var.gcs_service_account_secret_key
    }
  }
}
```

## Schema

### Required

- `bucket_name` (String) The name of the GCS bucket in which to store the logs.
- `name` (String) The name for the real-time logging configuration. Must be unique within the service.
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `authentication` (Attributes) Google Cloud Platform authentication credentials for GCS access. Provide either `account_name`, or `email` and `secret_key`. When this block is omitted entirely, defaults to the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME` (or `FASTLY_GCS_ACCOUNT_NAME`), `FASTLY_GCS_EMAIL`, and `FASTLY_GCS_SECRET_KEY` environment variables. (see [below for nested schema](#nestedatt--authentication))
- `compression_codec` (String) The codec used for compressing your logs. Valid values are `zstd`, `snappy`, and `gzip`. If the codec is `gzip`, `gzip_level` defaults to `3`; to use a different level, leave `compression_codec` unset and set `gzip_level` instead. Conflicts with `gzip_level`: setting both in the same request will result in an error.
- `format` (String) A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/).
- `format_version` (Number) The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.
- `gzip_level` (Number) The level of gzip encoding when sending logs. Valid values are `0` (no compression) through `9`. To compress at a specific gzip level, leave `compression_codec` unset and set this. Conflicts with `compression_codec`: setting both in the same request will result in an error.
- `message_type` (String) How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `classic`.
- `path` (String) The path to upload logs to. Must end with a trailing slash. If this field is left empty, the files will be saved in the bucket's root path.
- `period` (Number) How frequently log files are finalized so they can be available for reading, in seconds. Default `3600`.
- `placement` (String) Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.
- `processing_region` (String) The geographic region where the logs will be processed before streaming to Google Cloud Storage. Valid values are `us`, `eu`, and `none` for global. Default: `none`.
- `project_id` (String) Your Google Cloud Platform project ID. Not required if `account_name` is specified.
- `response_condition` (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- `timestamp_format` (String) `strftime`-specified timestamp format for log filename.

### Read-Only

- `id` (String) Terraform resource identifier.

<a id="nestedatt--authentication"></a>
### Nested Schema for `authentication`

Optional:

- `account_name` (String) The name of the Google Cloud Platform service account associated with the target log collection service. Not required if `email` and `secret_key` are provided. Can be set via the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME` environment variable (shared with Fastly's BigQuery and Pub/Sub logging endpoints), falling back to `FASTLY_GCS_ACCOUNT_NAME`.
- `email` (String, Sensitive) The `client_email` field in your service account authentication JSON. Not required if `account_name` is provided. Can be set via the `FASTLY_GCS_EMAIL` environment variable.
- `secret_key` (String, Sensitive) The `private_key` field in your service account authentication JSON. Not required if `account_name` is provided. Can be set via the `FASTLY_GCS_SECRET_KEY` environment variable.

## Import

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the endpoint from the Fastly API and
populate full state:

```shell
terraform import fastly_service_logging_gcs.example SERVICE_ID/VERSION/ENDPOINT_NAME
```

Example:

```shell
terraform import fastly_service_logging_gcs.example SU1Z0isxPaozGVKXdv0eY/3/gcs-example
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use
explicit service-version lifecycle actions to clone, validate, stage, or
activate a service version.

## Notes

- `authentication` groups credentials as the other logging endpoints do.
  Provide either `account_name` (the name of a Google Cloud IAM service
  account set up for [impersonation](https://www.fastly.com/documentation/guides/integrations/streaming-logs/configuring-google-iam-service-account-impersonation-for-fastly-logging/)),
  or `email` and `secret_key`. When the block is omitted entirely, defaults to
  the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME`, `FASTLY_GCS_EMAIL`, and
  `FASTLY_GCS_SECRET_KEY` environment variables — `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME`
  is shared with Fastly's BigQuery and Pub/Sub logging endpoints, since all
  three use the same Google Cloud service account. `email` and `secret_key`
  are sensitive and never appear in plan output. Once `account_name` is set,
  it can only be changed to a different value — not cleared back to unset —
  since the Fastly API rejects an explicit empty `account_name` on update.
  `account_name` also falls back to this resource's own `FASTLY_GCS_ACCOUNT_NAME`
  environment variable (used by the live provider) when
  `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME` is unset — no deprecation warning is
  emitted for this fallback, since `FASTLY_GCS_ACCOUNT_NAME` remains this
  resource's own, current variable name.
- `secret_key` must be a real PEM-encoded private key (PKCS8 or PKCS1) and
  must not contain leading or trailing whitespace — the Fastly API validates
  the credential and rejects both.
- `compression_codec` and `gzip_level` are mutually exclusive; the Fastly API
  rejects a request that sets both.
- `format` must produce valid JSON. If `format` is not sent, the API falls
  back to a general JSON log format similar to the one used by other
  streaming-logs integrations.
- Leaving `placement` unset is not the same as setting it to `none`: unset
  lets Fastly place the logging call automatically (`vcl_log` for
  `format_version` 2, `vcl_deliver` for `format_version` 1), while `none`
  suppresses the generated log statement entirely so you can write it
  yourself.
