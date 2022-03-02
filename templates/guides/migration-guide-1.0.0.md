---
page_title: "Upgrading to version 1"
subcategory: "Migration"
---

## Upgrading to version 1

**Resources omit `_v1` suffix**:

All resources now have a consistent naming convention which omit the previous v1 suffix.

- `fastly_service_v1` -> `fastly_service_vcl` (Fastly has two service types; this resource has always described the VCL one)
- `fastly_service_acl_entries_v1` -> `fastly_service_acl_entries`
- `fastly_service_dictionary_items_v1` -> `fastly_service_dictionary_items`
- `fastly_service_dynamic_snippet_content_v1` -> `fastly_service_dynamic_snippet_content`
- `fastly_user_v1` -> `fastly_user`

After upgrading, some existing projects might see an error such as:

```
Error: no schema available for module.fastly.<...> while reading state; this is a bug in Terraform and should be reported
```

This error occurs because the Terraform state file contains a different resource name to what's now exposed via the updated provider. To resolve this issue, remove the old resource from the state file and reimport the data using the new resource name.

As an example, if you had defined an `fastly_service_acl_entries_v1` resource, then you could remove and reimport using the following commands (any brackets `<...>` should be replaced with actual values):

```bash
terraform state rm 'fastly_service_acl_entries_v1.<your_resource_name>["<your_acl_name>"]'
terraform import 'fastly_service_acl_entries.<your_resource_name>["<your_acl_name>"]' <service_id>/<acl_id>

terraform state rm 'fastly_service_v1.<your_resource_name>'
terraform import 'fastly_service_vcl.<your_resource_name>' <service_id>
```

We recommend that you keep a backup of your state file so you can use it to reference any resource IDs necessary.

**Logging resources have consistent naming format**:

All logging resources now have a consistent naming convention based on the name of the logging provider, prefixed with `logging_`.

- `bigquerylogging` -> `logging_bigquery`
- `blobstoragelogging` -> `logging_blobstorage`
- `gcslogging` -> `logging_gcs`
- `httpslogging` -> `logging_https`
- `logentries` -> `logging_logentries`
- `papertrail` -> `logging_papertrail`
- `s3logging` -> `logging_s3`
- `splunk` -> `logging_splunk`
- `sumologic` -> `logging_sumologic`
- `syslog` -> `logging_syslog`

**Director `capacity` removed**:

The Fastly API never supported the `capacity` field for a `director` resource (this was added to the Terraform provider by mistake). Load balancing of director backends is managed by the `weight` field on each associated `backend` resource.

**GCS Logging field `email` renamed**:

The Fastly API was updated with a new [`user`](https://developer.fastly.com/reference/api/logging/gcs/) field to replace `email`.

**Logging `format` and `format_version` defaults changed**:

Pre-1.0.0 the default values for `format` and `format_version` were incorrectly set to an older version `1`. All new logging endpoints use the version `2` custom log format by default.

**Backend `auto_loadbalance` default changed**:

The Fastly web interface sets "Auto load balance" to "No" by default. The most common reason for having multiple backends in a single service is to route different paths to different backends, rather than load balance traffic between the backends. Conversely, pre-1.0.0, the terraform provider set`auto_loadbalance` to `true` by default, which was inconsistent and often unexpected. The default is now `false`.

**Gzip `content_types` and `extensions` type changed**:

The `content_types` and `extensions` fields for a `gzip` resource, pre-1.0.0, were implemented as a [`TypeSet`](https://www.terraform.io/plugin/sdkv2/schemas/schema-types#typeset) (an **unordered** collection of items whose state index is calculated by the hash of the attributes of the set). This would result in confusing and unexpected diffs. Now they are implemented as a [`TypeList`](https://www.terraform.io/plugin/sdkv2/schemas/schema-types#typelist) (an **ordered** collection of items).

**Automatically opt-in to `ignore_changes` behaviour for versionless resources**:

The versionless resources (ACL entries, Dictionary items and Dynamic VCL Snippets) are sometimes used in a way whereby they are "seeded" via Terraform and then updated/managed externally via the API or web interface. For this, the documentation suggests using `ignore_changes`, a built-in Terraform meta-argument, that allows the user to specify fields to ignore and from which to allow the state to drift.

However, sometimes this isn't obvious or the user doesn't understand this suggestion until it is too late, and data ends up getting lost. This happens because the user makes changes elsewhere and doesn't use `ignore_changes`, so Terraform takes action to remove the state drift and deletes their changes. This data is then unrecoverable.

Starting from version 1.0.0, Terraform now ignores any changes, and only allows the "dangerous" behaviour by explicitly opting in with a `manage_*` option (e.g. `manage_entries`, `manage_items`, `manage_snippets` depending on the versionless resource).
