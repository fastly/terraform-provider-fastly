# Terraform query support

## Purpose

This document explains how to use `terraform query` with the dual-model Fastly
Terraform provider.

`terraform query` is supported for the **explicit first-class resource family**
only. It is intended for read-only discovery and generated configuration.

It is not used for the compatibility resource family.

---

## Supported resource family

`terraform query` is supported for the **explicit first-class resource family**
only:

- `fastly_service_vcl_explicit`
- `fastly_service_domain_explicit`
- `fastly_service_backend_explicit`

These resources are first-class Terraform resources, so they can be discovered
independently and generated as separate resource blocks.

Example explicit configuration:

```hcl
resource "fastly_service_vcl_explicit" "example" {
  name = "example"
}

resource "fastly_service_domain_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = 1
  name       = "www.example.com"
}

resource "fastly_service_backend_explicit" "example" {
  service_id = fastly_service_vcl_explicit.example.id
  version    = 1
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}
```

---

## Compatibility resources use import instead

`terraform query` is not supported for the compatibility resource family:

- `fastly_service_vcl`

The compatibility resource uses nested configuration:

```hcl
resource "fastly_service_vcl" "example" {
  domain {
    name = "www.example.com"
  }

  backend {
    name    = "origin"
    address = "origin.example.com"
    port    = 443
  }
}
```

Because `fastly_service_vcl` owns nested configuration as one aggregate resource,
generated configuration for this resource family should come from Terraform
import, not from query:

```bash
terraform import fastly_service_vcl.example <service_id>
```

---

## Version selection

`terraform query` is read-only. It must not clone, activate, stage, or otherwise
mutate Fastly service versions.

For each Fastly service, query selects the version to read from using this order:

1. If the service has an active version, read from the active version.
2. If the service has no active version, read from the latest service version.

A Fastly service is expected to have at least one version because service
creation creates version `1`.

Generated explicit resources include the version number that was read.

If the generated version is active or locked, the generated configuration is
still useful for discovery and import. Before making changes with the explicit
resource family, clone or select a writable version and update the generated
resources to target that writable version.

### Query flow

```text
terraform query for explicit resources
  |
  v
inspect service S
  |
  | active version exists?
  |---- yes ---> read domains/backends from active version
  |
  |---- no ----> read domains/backends from latest version
  v
generate *_explicit resources with version pinned to the version read
  |
  v
no clone, no activation, no staging, no mutation
```

---

## Basic usage

Create a normal Terraform provider configuration.

```hcl
# main.tf
terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {}
```

Create a query configuration file:

```hcl
# fastly.tfquery.hcl
list "fastly_service_vcl_explicit" "all" {
  provider = fastly
}

list "fastly_service_domain_explicit" "all" {
  provider = fastly
}

list "fastly_service_backend_explicit" "all" {
  provider = fastly
}
```

The `provider` argument is required in every `list` block. It tells Terraform
which provider configuration to use.

Then run:

```bash
export FASTLY_API_KEY=<token>

terraform query
```

To generate Terraform configuration:

```bash
terraform query -generate-config-out=generated.tf
```

To produce machine-readable output:

```bash
terraform query -json
```

If `generated.tf` already exists, remove it before running
`terraform query -generate-config-out=generated.tf` again.

---

## Generated configuration

`terraform query -generate-config-out=generated.tf` produces Terraform resource
blocks and matching import blocks.

Example service output:

```hcl
resource "fastly_service_vcl_explicit" "my_service" {
  provider = fastly
  name     = "My Service"
}

import {
  to       = fastly_service_vcl_explicit.my_service
  provider = fastly
  identity = {
    service_id = "abc123"
  }
}
```

Example domain output:

```hcl
resource "fastly_service_domain_explicit" "www_example_com" {
  provider   = fastly
  service_id = "abc123"
  version    = 7
  name       = "www.example.com"
}

import {
  to       = fastly_service_domain_explicit.www_example_com
  provider = fastly
  identity = {
    service_id = "abc123"
    version    = 7
    name       = "www.example.com"
  }
}
```

Example backend output:

```hcl
resource "fastly_service_backend_explicit" "origin" {
  provider   = fastly
  service_id = "abc123"
  version    = 7
  name       = "origin"
  address    = "origin.example.com"
  port       = 443
}

import {
  to       = fastly_service_backend_explicit.origin
  provider = fastly
  identity = {
    service_id = "abc123"
    version    = 7
    name       = "origin"
  }
}
```

After generating configuration:

1. review the generated resources
2. copy the generated blocks into the desired Terraform configuration
3. run `terraform apply` to import the discovered resources into state
4. remove the import blocks after a successful import, or keep them as a
   historical record

---

## Editing generated explicit resources

Generated explicit resources are pinned to the version that query read.

If that version is active or locked, do not edit and apply those resources
directly. First clone or select a writable version, then update the generated
`version` arguments to target that writable version.

For example:

```hcl
resource "fastly_service_domain_explicit" "www_example_com" {
  service_id = fastly_service_vcl_explicit.my_service.id
  version    = var.service_version
  name       = "www.example.com"
}
```

The explicit resource family is intended for caller-managed lifecycle workflows.
Terraform writes to the version you specify; it does not choose, clone, or
activate versions during normal resource CRUD.

---

## Summary

```text
fastly_service_vcl
  -> compatibility nested resource family
  -> generated configuration through terraform import
  -> automatic clone and activation during CRUD

*_explicit resources
  -> explicit first-class resource family
  -> generated configuration through terraform query
  -> query reads active version, or latest version if no active version exists
  -> lifecycle is caller-managed
```
