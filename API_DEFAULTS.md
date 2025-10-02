# Handling API default values

The Fastly control plane API is, unfortunately, inconsistent in its
handling of default values for attributes of resources. Some of the
inconsistent behaviors generate unwelcome interactions with
Terraform's expectations of resource states, so this document provides
guidance for how to handle various scenarios when creating or
modifying a resource in the provider.

The goal of this guide is to eliminate, wherever possible, situations
where a user of the provider creates or modifies a resource and then
immediately sees 'intended changes' when executing `terraform
plan`. Users are frustrated and confused when this occurs, as they
don't understand why Terraform thinks additional changes are necessary
when their previous `terraform apply` process completed successfully
and they haven't made changes to the resources outside of Terraform.

# Type A

* The attribute has a documented default.

* The default is applied when the user does not provide a value, or
  provides a value of `null`, for the attribute in a `POST` operation
  (creating the resource), and when the user sets the attribute's
  value to `null` in a `PUT` or `PATCH` operation (updating the
  resource).

* When the default is in effect, the attribute *does not appear* in
  the response bodies returned by `GET`, `PUT`, and `PATCH`
  operations.

* When the user provides a value for the attribute which happens to
  match the default value, that value *does appear* in the response
  bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Examples

## Terraform expectations

This type matches Terraform's expectations.

# Type B

* The attribute has a documented default.

* The default is applied when the user does not provide a value, or
  provides a value of `null`, for the attribute in a `POST` operation
  (creating the resource), and when the user sets the attribute's
  value to `null` in a `PUT` or `PATCH` operation (updating the
  resource).

* When the default is in effect, the attribute *does not appear* in
  the response bodies returned by `GET`, `PUT`, and `PATCH`
  operations.

* When the user provides a value for the attribute which happens to
  match the default value, that value *does not appear* in the
  response bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Examples

## Terraform expectations

The last two items in this type does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be missing. When the user executes `terraform plan` after any of these
operations, Terraform will propose to set the value again, as it
believes the value was unset/removed in the resource being managed.

# Type C

* The attribute has a documented default.

* The default is applied when the user does not provide a value, or
  provides a value of `null`, for the attribute in a `POST` operation
  (creating the resource), and when the user sets the attribute's
  value to `null` in a `PUT` or `PATCH` operation (updating the
  resource).

* The attribute *always appears* in the response bodies returned by
  `GET`, `PUT`, and `PATCH` operations.

## Examples

## Terraform expectations

The last item in this type does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be present even if the user did not include it in their HCL. When the
user executes `terraform plan` after any of these operations,
Terraform will propose to remove the value, as it believes the value
was set outside of Terraform in the resource being managed.

# Type D

* The attribute has a documented default.

* The default is applied when the user does not provide a value, or
  provides a value of `null`, for the attribute in a `POST` operation
  (creating the resource), and when the user sets the attribute's
  value to `null` in a `PUT` or `PATCH` operation (updating the
  resource).

* When the default is in effect, the attribute *does appear* in the
  response bodies returned by `GET`, `PUT`, and `PATCH` operations,
  however its value is the 'base' value for the attribute's type
  (empty string for string attributes, zero for numeric attributes).

* When the user provides a value for the attribute which happens to
  match the default value, that value *does appear* in the response
  bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Examples

## Terraform expectations

The second-to-last item in this type does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be present even if the user did not include it in their HCL. When the
user executes `terraform plan` after any of these operations,
Terraform will propose to remove the value, as it believes the value
was set outside of Terraform in the resource being managed.

# Type E

* The attribute has a documented default.

* The attribute is not optional, it must always be included in the
  request body in `POST` operations (creating the resource), and
  `PUT`, and `PATCH` operations (updating the resource).

* The API accepts the 'base' value for the attribute's type (empty
  string for string attributes, zero for numeric attributes) but
  treats that as equivalent to the default value.

* The attribute *always appears* in the response bodies returned by
  `GET`, `PUT`, and `PATCH` operations. The returned value is the
  value provided by the user.

## Examples

* The `period` attribute in HTTPS Logging endpoints.

## Terraform expectations

This type matches Terraform's expectations.

# Type F

* There are two attributes with documented defaults.

* Both attributes are optional, and the API is documented to consider
  them mutually exclusive (it is documented to return an error if both
  attributes are included in the request body of a `POST`, `PUT`, or
  `PATCH` operation).

* When a value for one attribute is provided in the request body of a
  `POST` operation (creating the resource), the API may set the other
  attribute to a default value derived from the provided value.

* When a value for one attribute is provided in the request body of a
  `PUT` or `PATCH` operation (updating the resource), the API may
  change the value of the other attribute to a value derived from the
  provided value.

* The attributes *always appear* in the response bodies returned by
  `GET`, `PUT`, and `PATCH` operations.

## Examples

* The `compression_codec` and `gzip_level` attributes in various
  Logging endpoints (HTTPS, S3, and others).

## Terraform expectations

The third and fourth items in this type do not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), one or both of the
attributes will be present even if the user did not include them in
their HCL. In addition, the value of one of the attributes may have
changed since the last time it was read. When the user executes
`terraform plan` after any of these operations, Terraform will propose
to change or remove the value, as it believes the value was modified
(outside of Terraform) in the resource being managed.
