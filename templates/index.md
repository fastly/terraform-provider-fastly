---
page_title: "fastly Provider"
subcategory: ""
description: |-
---

# Fastly Provider

The Fastly provider is used to interact with the content delivery network (CDN)
provided by Fastly.

In order to use this Provider, you must have an active account with Fastly.
Pricing and signup information can be found at https://www.fastly.com/signup

Use the navigation to the left to read about the available resources.

The Fastly provider prior to version v0.13.0 requires using
[--parallelism=1](/docs/commands/apply.html#parallelism-n) for `apply` operations.

## Example Usage

```hcl
# Configure the Fastly Provider
provider "fastly" {
  api_key = "test"
}

# Create a Service
resource "fastly_service_v1" "myservice" {
  name = "myawesometestservice"

  # ...
}
```

## Authentication

The Fastly provider offers an API key based method of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

- Static API key
- Environment variables


### Static API Key

Static credentials can be provided by adding a `api_key` in-line in the
Fastly provider block:

Usage:

```hcl
provider "fastly" {
  api_key = "test"
}

resource "fastly_service_v1" "myservice" {
  # ...
}
```

You can create a credential on the Personal API Tokens page: https://manage.fastly.com/account/personal/tokens

### Environment variables

You can provide your API key via `FASTLY_API_KEY` environment variable,
representing your Fastly API key. When using this method, you may omit the
Fastly `provider` block entirely:

```hcl
resource "fastly_service_v1" "myservice" {
  # ...
}
```

Usage:

```
$ export FASTLY_API_KEY="afastlyapikey"
$ terraform plan
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `api_key` - (Optional) This is the API key. It must be provided, but
  it can also be sourced from the `FASTLY_API_KEY` environment variable

* `base_url` - (Optional) This is the API server hostname. It is required
  if using a private instance of the API and otherwise defaults to the
  public Fastly production service. It can also be sourced from the
  `FASTLY_API_URL` environment variable
