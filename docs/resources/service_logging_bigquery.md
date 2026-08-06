---
page_title: "fastly_service_logging_bigquery Resource - fastly"
subcategory: ""
description: |-
  Fastly service BigQuery logging endpoint resource. Writes directly to the specified writable service version.
---

# fastly_service_logging_bigquery (Resource)

Fastly service BigQuery logging endpoint resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a BigQuery real-time logging endpoint on the configured service
version. It does not clone, activate, or stage service versions.

To have the provider manage the version lifecycle for you instead, use the
nested `logging_bigquery` block on `fastly_service_cdn_auto` or
`fastly_service_compute_auto` — see "Automatic-lifecycle usage" below.

## Example Usage

```terraform
resource "fastly_service_logging_bigquery" "example" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "bigquery-example"

  project_id = "my-gcp-project"
  dataset    = "my_dataset"
  table      = "my_table"

  authentication = {
    email      = var.bigquery_service_account_email
    secret_key = var.bigquery_service_account_secret_key
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
resource "fastly_service_logging_bigquery" "linked_account" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "bigquery-linked-account"

  project_id = "my-gcp-project"
  dataset    = "my_dataset"
  table      = "my_table"
  template   = "_%Y%m%d"

  authentication = {
    account_name = "my-linked-service-account"
  }
  processing_region = "eu"

  format             = "{\n \"timestamp\":\"%{begin:%Y-%m-%dT%H:%M:%S}t\"\n}"
  format_version     = 2
  placement          = "none"
  response_condition = fastly_service_condition.errors_only.name
}
```

Attaching to a Compute service — the VCL-only attributes must be omitted:

```terraform
resource "fastly_service_logging_bigquery" "compute" {
  service_id = fastly_service_compute.example.id
  version    = 1
  name       = "bigquery-compute"

  project_id = "my-gcp-project"
  dataset    = "my_dataset"
  table      = "my_table"

  authentication = {
    email      = var.bigquery_service_account_email
    secret_key = var.bigquery_service_account_secret_key
  }
}
```

## Automatic-lifecycle usage

Inside the `_auto` service resources, BigQuery logging is a nested block and
the provider clones, validates, and activates a new service version whenever
the block changes. The nested block takes the same arguments as this
resource, minus `service_id` and `version`, which the parent service owns.

```terraform
resource "fastly_service_cdn_auto" "example" {
  name = "my-service"

  domain {
    name = "www.example.com"
  }

  logging_bigquery {
    name       = "bigquery-example"
    project_id = "my-gcp-project"
    dataset    = "my_dataset"
    table      = "my_table"

    authentication = {
      email      = var.bigquery_service_account_email
      secret_key = var.bigquery_service_account_secret_key
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

  logging_bigquery {
    name       = "bigquery-compute"
    project_id = "my-gcp-project"
    dataset    = "my_dataset"
    table      = "my_table"

    authentication = {
      email      = var.bigquery_service_account_email
      secret_key = var.bigquery_service_account_secret_key
    }
  }
}
```

## Schema

### Required

- `dataset` (String) The ID of your BigQuery dataset.
- `name` (String) The name for the real-time logging configuration. Must be unique within the service.
- `project_id` (String) The ID of your Google Cloud Platform project.
- `service_id` (String) Fastly service ID.
- `table` (String) The ID of your BigQuery table.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `authentication` (Attributes) Google Cloud Platform authentication credentials for BigQuery access. Provide either `account_name`, or `email` and `secret_key`. When this block is omitted entirely, defaults to the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME`, `FASTLY_BQ_EMAIL`, and `FASTLY_BQ_SECRET_KEY` environment variables. (see [below for nested schema](#nestedatt--authentication))
- `format` (String) A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/). Must produce valid JSON that matches the schema of your BigQuery table.
- `format_version` (Number) The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if format_version is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.
- `placement` (String) Where in the generated VCL the logging call should be placed. If not set, endpoints with format_version of 2 are placed in vcl_log and those with format_version of 1 are placed in vcl_deliver. Valid value is `none`.
- `processing_region` (String) The geographic region where the logs will be processed before streaming. Valid values are `us`, `eu`, and `none` for global. Default: `none`.
- `response_condition` (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- `template` (String) A template string used to generate a BigQuery table name suffix.

### Read-Only

- `id` (String) Terraform resource identifier.

<a id="nestedatt--authentication"></a>
### Nested Schema for `authentication`

Optional:

- `account_name` (String) The name of the Google Cloud IAM service account used for impersonation-based authentication, associated with the target log collection service. Not required if `email` and `secret_key` are provided. Can be set via the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME` environment variable, shared with Fastly's GCS and Pub/Sub logging endpoints.
- `email` (String, Sensitive) The `client_email` field in your service account authentication JSON. Not required if `account_name` is provided. Can be set via the `FASTLY_BQ_EMAIL` environment variable.
- `secret_key` (String, Sensitive) The `private_key` field in your service account authentication JSON. Not required if `account_name` is provided. Can be set via the `FASTLY_BQ_SECRET_KEY` environment variable.

## Import

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the endpoint from the Fastly API and
populate full state:

```shell
terraform import fastly_service_logging_bigquery.example SERVICE_ID/VERSION/ENDPOINT_NAME
```

Example:

```shell
terraform import fastly_service_logging_bigquery.example SU1Z0isxPaozGVKXdv0eY/3/bigquery-example
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
  the `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME`, `FASTLY_BQ_EMAIL`, and
  `FASTLY_BQ_SECRET_KEY` environment variables — `FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME`
  is shared with Fastly's GCS and Pub/Sub logging endpoints, since all three use
  the same Google Cloud service account. `email` and `secret_key` are
  sensitive and never appear in plan output. Once `account_name` is set, it
  can only be changed to a different value — not cleared back to unset —
  since the Fastly API rejects an explicit empty `account_name` on update.
- `secret_key` must be a real PEM-encoded private key (PKCS8 or PKCS1) and
  must not contain leading or trailing whitespace — the Fastly API validates
  the credential and rejects both.
- `format` must produce valid JSON that matches the schema of your BigQuery
  table. If `format` is not sent, the API falls back to a general JSON log
  format similar to the one used by other streaming-logs integrations.
- Leaving `placement` unset is not the same as setting it to `none`: unset
  lets Fastly place the logging call automatically (`vcl_log` for
  `format_version` 2, `vcl_deliver` for `format_version` 1), while `none`
  suppresses the generated log statement entirely so you can write it
  yourself.
