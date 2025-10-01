# Handling API default values

The Fastly control plane API is, unfortunately, inconsistent in its
handling of default values for attributes of resources. Some of the
inconsistent behaviors generate unwelcome interactions with
Terraform's expectations of resource states, so this document provides
guidance for how to handle various scenarios when creating or
modifying a resource in the provider.

# Scenario A

* The attribute has a documented default.

* That default is applied when the user does not provide a value for
  the attribute in a `POST` operation (creating the resource), and
  when the user sets the attribute's value to `null` in a `PUT` or
  `PATCH` operation (updating the resource).

* When the default is in effect, the attribute *does not appear* in
  the response bodies returned by `GET`, `PUT`, and `PATCH`
  operations.

* When the user provides a value for the attribute which happens to
  match the default value, that value *does appear* in the response
  bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Terraform results

This scenario matches Terraform's expectations.

# Scenario B

* The attribute has a documented default.

* That default is applied when the user does not provide a value for
  the attribute in a `POST` operation (creating the resource), and
  when the user sets the attribute's value to `null` in a `PUT` or
  `PATCH` operation (updating the resource).

* When the default is in effect, the attribute *does not appear* in
  the response bodies returned by `GET`, `PUT`, and `PATCH`
  operations.

* When the user provides a value for the attribute which happens to
  match the default value, that value *does not appear* in the
  response bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Terraform results

The last item in this scenario does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be missing. When the user executes `terraform plan` after any of these
operations, Terraform will propose to set the value again, as it
believes the value was unset/removed in the resource being managed.

# Scenario C

* The attribute has a documented default.

* That default is applied when the user does not provide a value for
  the attribute in a `POST` operation (creating the resource), and
  when the user sets the attribute's value to `null` in a `PUT` or
  `PATCH` operation (updating the resource).

* The attribute *always appears* in the response bodies returned by
  `GET`, `PUT`, and `PATCH` operations.

## Terraform results

The last item in this scenario does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be present even if the user did not include it in their HCL. When the
user executes `terraform plan` after any of these operations,
Terraform will propose to remove the value, as it believes the value
was set in the resource being managed.

# Scenario D

* The attribute has a documented default.

* That default is applied when the user does not provide a value for
  the attribute in a `POST` operation (creating the resource), and
  when the user sets the attribute's value to `null` in a `PUT` or
  `PATCH` operation (updating the resource).

* When the default is in effect, the attribute *does appear* in the
  response bodies returned by `GET`, `PUT`, and `PATCH` operations,
  however its value is the 'base' value for the attribute's type
  (empty string for string attributes, zero for numeric attributes).

* When the user provides a value for the attribute which happens to
  match the default value, that value *does appear* in the response
  bodies returned by `GET`, `PUT`, and `PATCH` operations.

## Terraform results

The second-to-last item in this scenario does not match Terraform's
expectations: when the resource is read from the API (during import,
during planning, or after creation/modification), the attribute will
be present even if the user did not include it in their HCL. When the
user executes `terraform plan` after any of these operations,
Terraform will propose to remove the value, as it believes the value
was set in the resource being managed.

# Scenario E

* The attribute has a documented default.

* The attribute is not optional, it must always be included in the
  request body in `POST`, `PUT`, and `PATCH` operations.

* The API accepts the 'base' value for the attribute's type (empty
  string for string attributes, zero for numeric attributes) but
  treats that as equivalent to the default value.

* The attribute *always appears* in the response bodies returned by
  `GET`, `PUT`, and `PATCH` operations.

## Terraform results

This scenario matches Terraform's expectations.
