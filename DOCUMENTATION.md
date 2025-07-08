# Provider Documentation

## Building and Reviewing

The documentation is built from templates stored in the `templates`
folder, along with examples stored in the `examples` folder. All of
this content is copied to the `docs` folder, with the templates fully
rendered, when the documentation is built.

To validate the `templates` directory structure and its references to
schemas in the provider and to examples, run:

```shell
make validate-docs
```

To build the documentation, replacing all of the contents of the
`docs` folder:

```shell
make generate-docs
```

To preview a single documentation file, copy the content of the
generated Markdown file into the [preview
tool](https://registry.terraform.io/tools/doc-preview) on the
HashiCorp website.

## Templates

Every resource and data source in the provider has its own
documentation, which includes details of making use of that
resource/data source, along with the details of its schema.

If you do not write a template for a resource or data source that you
create, `make generate-docs` will use a basic template to generate the
documentation for the provider. This template will include only the
name of the resource/data source and the details of its schema. For
resources in particular, this is insufficient to guide users as it
does not include examples of HCL which make use of the resource, or
instructions for how to import an existing resource into the Terraform
state.

Note: The documentation generation tool gathers the list of known data
sources and resources from the provider itself, not by scanning the
source directory. As a result, if you have created a new data source
or resource but you have not added into the top-level schema in
[provider.go](fastly/provider.go) it will not be included in the
generated documentation. If you create a template for the new data
source or resource and run `make validate-docs` the validator will
produce an error as it will be unable to find the new data source or
resource. The simplest way to avoid this problem is to write and
execute the acceptance tests for the new data source or resource
*before* working on its documentation; in order for the acceptance
tests to pass the new data source or resource must already be included
in the top-level schema.

Before submitting a PR with your new data source or resource, you
should add a proper template for it, and also ensure that the schema
within the data source or resource has been properly documented as
well. Review the following three sections for details.

## Data Sources

Each data source should have a template file and an example file
showing basic usage of the data source in HCL. The template for [NGWAF
Workspaces](templates/data-sources/ngwaf_workspaces.md.tmpl) can be
used as a reference.

## Resources

Each resource should have a template file and at least two example
files: one showing basic usage of the resource in HCL, and the
other demonstrating how to import an existing resource into the
Terraform state. The examples should be fully functional: they should
include all resources necessary to create the resource being
documented. The template for [Config
Stores](templates/resources/configstore.md.tmpl) can be used as a
reference.

## Schema Attributes

Every attribute in every schema must include a description. The
description should provide all of the information necessary for a user
to understand the purpose and usage of that attribute.

To write consistent and helpful attribute descriptions, follow these steps:

1. Use the description (from the public documentation) of the
   corresponding attribute of the API resource that the schema
   represents, whenever possible.

1. If the attribute allows only a limited set of values (an
   enumeration), include a sentence such as "Accepted values are
   `value 1`, `value 2`, and `value 3`".

1. If the attribute allows only a range of values (numeric), include a
   sentence such as "Minimum 1 and maximum 10,000".

1. If the attribute has a default value, include a sentence such as
   "Default value `value`".

1. If the attribute's value is an ID which can be obtained from
   another data source or resource, include a sentence such as "Can be
   obtained from `fastly_service_vcl`, `fastly_service_compute`, or
   `fastly_services`".

1. There is no need to include information about the attribute being
   optional, computed, or read-only, as the schema documentation
   generator will include that information in the documentation of the
   attribute.

Composing a description by following these steps, in this order, will
result in a consistent presentation for the user of the data source or
resource.
