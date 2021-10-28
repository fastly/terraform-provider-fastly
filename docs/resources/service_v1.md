---
layout: "fastly"
page_title: "Fastly: service_v1"
sidebar_current: "docs-fastly-resource-service-v1"
description: |-
  Provides an Fastly Service
---

# fastly_service_v1

Provides a Fastly Service, representing the configuration for a website, app,
API, or anything else to be served through Fastly. A Service encompasses Domains
and Backends.

The Service resource requires a domain name that is correctly set up to direct
traffic to the Fastly service. See Fastly's guide on [Adding CNAME Records][fastly-cname]
on their documentation site for guidance.

## Example Usage

Basic usage:

```terraform
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}
```

Basic usage with an Amazon S3 Website and that removes the `x-amz-request-id` header:

```terraform
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address       = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name          = "AWS S3 hosting"
    port          = 80
    override_host = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
  }

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  gzip {
    name          = "file extensions and content types"
    extensions    = ["css", "js"]
    content_types = ["text/html", "text/css"]
  }

  force_destroy = true
}

resource "aws_s3_bucket" "website" {
  bucket = "demo.notexample.com"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
```

Basic usage with [custom
VCL](https://docs.fastly.com/vcl/custom-vcl/uploading-custom-vcl/):

```terraform
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true

  vcl {
    name    = "my_custom_main_vcl"
    content = "${file("${path.module}/my_custom_main.vcl")}"
    main    = true
  }

  vcl {
    name    = "my_custom_library_vcl"
    content = "${file("${path.module}/my_custom_library.vcl")}"
  }
}
```

Basic usage with [custom Director](https://developer.fastly.com/reference/api/load-balancing/directors/director/):

```terraform
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  backend {
    address = "127.0.0.2"
    name    = "origin2"
    port    = 80
  }

  director {
    name = "mydirector"
    quorum = 0
    type = 3
    backends = [ "origin1", "origin2" ]
  }

  force_destroy = true
}
```

-> **Note:** The following example is only available from 0.20.0 of the Fastly Terraform provider.

Basic usage with [Web Application Firewall](https://developer.fastly.com/reference/api/waf/):

```terraform
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "example.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "WAF_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # This condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic.
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name              = "WAF_Response"
    status            = "403"
    response          = "Forbidden"
    content_type      = "text/html"
    content           = "<html><body>Forbidden</body></html>"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "WAF_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}
```

-> **Note:** For an AWS S3 Bucket, the Backend address is
`<domain>.s3-website-<region>.amazonaws.com`. The `override_host` attribute
should be set to `<bucket_name>.s3-website-<region>.amazonaws.com` in the `backend` block. See the
Fastly documentation on [Amazon S3][fastly-s3].

[fastly-cname]: https://docs.fastly.com/en/guides/adding-cname-records
[fastly-conditionals]: https://docs.fastly.com/en/guides/using-conditions
[fastly-sumologic]: https://developer.fastly.com/reference/api/logging/sumologic/
[fastly-gcs]: https://developer.fastly.com/reference/api/logging/gcs/

## Import

Fastly Services can be imported using their service ID, e.g.

```txt
$ terraform import fastly_service_v1.demo xxxxxxxxxxxxxxxxxxxx
```

By default, either the active version will be imported, or the latest version if no version is active.
Alternatively, a specific version of the service can be selected by appending an `@` followed by the version number to the service ID, e.g.

```txt
$ terraform import fastly_service_v1.demo xxxxxxxxxxxxxxxxxxxx@2
```