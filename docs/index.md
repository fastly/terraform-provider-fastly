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
[--parallelism=1](https://developer.hashicorp.com/terraform/cli/commands/apply#parallelism-n) for `apply` operations.

## Example Usage

```terraform
# Terraform 0.13+ requires providers to be declared in a "required_providers" block
terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = ">= 7.1.0"
    }
  }
}

# Configure the Fastly Provider
provider "fastly" {
  api_key = "test"
}

# Create a Service
resource "fastly_service_vcl" "myservice" {
  name = "myawesometestservice"

  # ...
}
```

## Importing

Importing using the standard Terraform documentation should work for most use cases. If any of your Fastly resources have attributes which are considered sensitive (e.g. credentials for logging endpoints, TLS private keys) please see the Sensitive Attributes guide for the configuration necessary to ensure that those attributes will be imported.


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

```terraform
provider "fastly" {
  api_key = "test"
}

resource "fastly_service_vcl" "myservice" {
  # ...
}
```

You can create a credential on the Personal API Tokens page: https://manage.fastly.com/account/personal/tokens

### Environment variables

You can provide your API key via `FASTLY_API_KEY` environment variable,
representing your Fastly API key. When using this method, you may omit the
Fastly `provider` block entirely:

```terraform
resource "fastly_service_vcl" "myservice" {
# ...
}
```

Usage:

```sh
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

* `no_auth` - (Optional) Set to `true` if your configuration only consumes data sources that do not require authentication, such as `fastly_ip_ranges`. Default: `false`

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String) Fastly API Key from https://app.fastly.com/#account
- `base_url` (String) Fastly API URL
- `force_http2` (Boolean) Set this to `true` to disable HTTP/1.x fallback mechanism that the underlying Go library will attempt upon connection to `api.fastly.com:443` by default. This may slightly improve the provider's performance and reduce unnecessary TLS handshakes. Default: `false`
- `no_auth` (Boolean) Set to `true` if your configuration only consumes data sources that do not require authentication, such as `fastly_ip_ranges`
