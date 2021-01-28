---
layout: "fastly"
page_title: "Fastly: service_dynamic_snippet_content_v1"
sidebar_current: "docs-fastly-resource-service-dynamic-snippet-content-v1"
description: |-
  Provides a means to define blocks of VCL logic that is inserted into your service through Fastly dynamic snippets.
---

# fastly_service_dynamic_snippet_content_v1

Defines content that represents blocks of VCL logic that is inserted into your service.  This resource will populate the content of a dynamic snippet and allow it to be manged without the creation of a new service verison. 
 
~> **Warning:** Terraform will take precedence over any changes you make through the API. Such changes are likely to be reversed if you run Terraform again.  

If Terraform is being used to populate the initial content of a dynamic snippet which you intend to manage via the API, then the lifecycle `ignore_changes` field can be used with the resource.  An example of this configuration is provided below.


## Example Usage (Terraform >= 0.12.6)

Basic usage:

```hcl 
resource "fastly_service_v1" "myservice" {
  name = "snippet_test"

  domain {
    name    = "snippet.fastlytestdomain.com"
    comment = "snippet test"
  }

  backend {
    address = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet"
    type     = "recv"
    priority = 110
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content" {
  for_each = {
    for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

}
```

Multiple dynamic snippets:

```hcl
resource "fastly_service_v1" "myservice" {
  name = "snippet_test"

  domain {
    name    = "snippet.fastlytestdomain.com"
    comment = "snippet test"
  }

  backend {
    address = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet One"
    type     = "recv"
    priority = 110
  }

  dynamicsnippet {
       name     = "My Dynamic Snippet Two"
       type     = "recv"
       priority = 110
     }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content_one" {
  for_each = {
    for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet One"
  }

  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header-one = \"true\";\n}"

}

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content_two" {
  for_each = {
    for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet Two"
  }

  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header-two = \"true\";\n}"

}
```

## Example Usage (Terraform >= 0.12.0 && &lt; 0.12.6)

`for_each` attributes were not available in Terraform before 0.12.6, however, users can still use `for` expressions to achieve
similar behaviour as seen in the example below.

~> **Warning:** Terraform might not properly calculate implicit dependencies on computed attributes when using `for` expressions

For scenarios such as adding a Dynamic Snippet to a service and at the same time, creating the Dynamic Snippets (`fastly_service_dynamic_snippet_content_v1`)
resource, Terraform will not calculate implicit dependencies correctly on `for` expressions. This will result in index lookup
problems and the execution will fail.

For those scenarios, it's recommended to split the changes into two distinct steps:

1. Add the `dynamicsnippet` block to the `fastly_service_v1` and apply the changes
2. Add the `fastly_service_dynamic_snippet_content_v1` resource with the `for` expressions to the HCL and apply the changes

Usage:

```hcl
resource "fastly_service_v1" "myservice" {
  ...
  dynamicsnippet {
    name     = "My Dynamic Snippet"
    type     = "recv"
    priority = 110
  }
  ...
}

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content" {
  service_id = fastly_service_v1.myservice.id
  snippet_id = { for s in fastly_service_v1.myservice.dynamicsnippet : s.name => s.snippet_id }["My Dynamic Snippet"]

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

}
```

### Supporting API dynamic snippet updates with ignore_changes

The following example demonstrates how the lifecycle `ignore_changes` field can be used to suppress updates against the 
content in a dynamic snippet.  If, after your first deploy, the Fastly API is to be used to manage items in a dynamic snippet, then this will stop Terraform realigning the remote state with the initial content defined in your HCL.

```hcl
...

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content" {
  for_each = {
    for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

  lifecycle {
    ignore_changes = [content, ]
  }
}
```

## Attributes Reference

* [fastly-vcl](https://developer.fastly.com/reference/api/vcl-services/vcl/)
* [fastly-vcl-snippets](https://developer.fastly.com/reference/api/vcl-services/snippet/)

## Import

This is an example of the import command being applied to the resource named `fastly_service_dynamic_snippet_content_v1.content`
The resource ID is a combined value of the `service_id` and `snippet_id` separated by a forward slash.

```
$ terraform import fastly_service_dynamic_snippet_content_v1.content xxxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxx
```

If Terraform is already managing remote content against a resource being imported then the user will be asked to remove it from the existing Terraform state.
The following is an example of the Terraform state command to remove the resource named `fastly_service_dynamic_snippet_content_v1.content` from the Terraform state file.

```
$ terraform state rm fastly_service_dynamic_snippet_content_v1.content
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **content** (String) The VCL code that specifies exactly what the snippet does
- **service_id** (String) The ID of the service that the dynamic snippet belongs to
- **snippet_id** (String) The ID of the dynamic snippet that the content belong to

### Optional

- **id** (String) The ID of this resource.
