---
page_title: "fastly_service_backend Resource - fastly"
subcategory: ""
description: |-
  Fastly service backend resource. Writes directly to the specified writable service version.
---

# fastly_service_backend (Resource)

Fastly service backend resource. Writes directly to the specified writable service version.

This resource is part of the explicit/default first-class resource family. It
manages a backend on the configured service version. It does not clone, activate,
or stage service versions.

## Example Usage

```terraform
resource "fastly_service_backend" "origin" {
  service_id = fastly_service_cdn.example.id
  version    = 1
  name       = "origin"

  address = "api.example.com"
  port    = 443
  use_ssl = true
}
```

## Schema

### Required

- `address` (String) An IPv4 address, IPv6 address, or hostname for the backend.
- `name` (String) Name for this backend. Must be unique within the service.
- `service_id` (String) Fastly service ID.
- `version` (Number) Writable Fastly service version to modify.

### Optional

- `auto_loadbalance` (Boolean) Whether this backend should be included in automatic load balancing. CDN services only. Default `false`.
- `between_bytes_timeout` (Number) How long to wait between bytes in milliseconds. Default `10000`.
- `comment` (String) Optional comment for the backend.
- `connect_timeout` (Number) How long to wait for a timeout in milliseconds. Default `1000`.
- `error_threshold` (Number) Number of errors to allow before the backend is marked as down. Default `0`.
- `first_byte_timeout` (Number) How long to wait for the first byte in milliseconds. Default `15000`.
- `healthcheck` (String) Name of a defined healthcheck to assign to this backend.
- `keepalive_time` (Number) How long in seconds to keep a persistent connection to the backend between requests.
- `max_conn` (Number) Maximum number of connections for this backend. Default `200`.
- `max_lifetime` (Number) Maximum time from creation, in milliseconds, that a pooled HTTP keepalive connection is eligible for reuse. `0` is treated as unlimited.
- `max_tls_version` (String) Maximum allowed TLS version on SSL connections to this backend.
- `max_use` (Number) Maximum number of requests allowed over a single pooled HTTP keepalive connection. `0` is treated as unlimited.
- `min_tls_version` (String) Minimum allowed TLS version on SSL connections to this backend.
- `override_host` (String) Hostname to override the Host header.
- `port` (Number) The port number on which the backend responds. Default `80`.
- `prefer_ipv6` (Boolean) Prefer IPv6 connections to origins for hostname backends. Default `false` for CDN services.
- `request_condition` (String) Name of a request condition which, if met, selects this backend.
- `share_key` (String) Value that, when shared across backends, enables those backends to share the same health check.
- `shield` (String) POP of the shield designated to reduce inbound load.
- `ssl_ca_cert` (String) CA certificate attached to origin.
- `ssl_cert_hostname` (String) Hostname used for certificate validation. Does not affect SNI.
- `ssl_check_cert` (Boolean) Whether to strictly check SSL certificates. Default `true`.
- `ssl_ciphers` (String) Cipher list for TLS connections to this backend.
- `ssl_client_cert` (String, Sensitive) Client certificate used when connecting to the backend.
- `ssl_client_key` (String, Sensitive) Client key used when connecting to the backend.
- `ssl_sni_hostname` (String) Hostname used for SNI in the TLS handshake.
- `use_ssl` (Boolean) Whether to use SSL to reach the backend. Default `false`.
- `weight` (Number) Portion of traffic to send to this backend. Default `100`.

### Read-Only

- `id` (String) Terraform resource identifier.

## Import

`fastly_service_backend` has a stable Framework identity of `service_id + name`.
The `version` argument is not part of the stable identity because explicit
resources can move the same logical backend from one service version to another.

For import-from-scratch with the Terraform CLI, include the service version in
the import ID so the provider can read the backend from the Fastly API and
populate full state:

```shell
terraform import fastly_service_backend.origin SERVICE_ID/VERSION/BACKEND_NAME
```

Example:

```shell
terraform import fastly_service_backend.origin SU1Z0isxPaozGVKXdv0eY/3/origin
```

You can also use Terraform's identity-based import flow with the stable identity
fields. The resource configuration must still provide the `version` argument
because the provider needs a service version to read the backend from Fastly:

```terraform
resource "fastly_service_backend" "origin" {
  service_id = "SU1Z0isxPaozGVKXdv0eY"
  version    = 3
  name       = "origin"

  address = "api.example.com"
  port    = 443
}

import {
  to = fastly_service_backend.origin

  identity = {
    service_id = "SU1Z0isxPaozGVKXdv0eY"
    name       = "origin"
  }
}
```

## Version lifecycle

This resource does not clone, activate, or stage service versions. Use explicit
service-version lifecycle actions to clone, validate, stage, or activate a
service version.
