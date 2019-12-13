## 0.11.1 (Unreleased)

BUG FIXES:

* data-source/fastly_ip_ranges: Use `go-fastly` client in order to fetch Fastly's assigned IP ranges ([#201](https://github.com/terraform-providers/terraform-provider-fastly/pull/201))

## 0.11.0 (October 15, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1/dictionary: Add `write_only` argument ([#189](https://github.com/terraform-providers/terraform-provider-fastly/pull/189))

NOTES:

* provider: The underlying Terraform codebase dependency for the provider SDK and acceptance testing framework has been migrated from `github.com/hashicorp/terraform` to `github.com/hashicorp/terraform-plugin-sdk`. They are functionality equivalent and this should only impact codebase development to switch imports. For more information see the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html). ([#191](https://github.com/terraform-providers/terraform-provider-fastly/pull/191))
* provider: The actual Terraform version used by the provider will now be included in the `User-Agent` header for Terraform 0.12 and later. Terraform 0.11 and earlier will use `Terraform/0.11+compatible` as this information was not accessible in those versions. ([#182](https://github.com/terraform-providers/terraform-provider-fastly/pull/182))

## 0.10.0 (October 02, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `cloned_version` argument ([#190](https://github.com/terraform-providers/terraform-provider-fastly/pull/190))

## 0.9.0 (August 07, 2019)

FEATURES:

* **New Resource:** `fastly_service_acl_entries_v1` ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))
* **New Resource:** `fastly_service_dictionary_items_v1` ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))
* **New Resource:** `fastly_service_dynamic_snippet_content_v1` ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))

ENHANCEMENTS:

* resource/fastly_service_v1: Add `acl` argument ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))
* resource/fastly_service_v1: Add `dictionary` argument ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))
* resource/fastly_service_v1: Add `dynamicsnippet` argument ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))

NOTES:

* provider: Update `go-fastly` client to v1.2.1 ([#184](https://github.com/terraform-providers/terraform-provider-fastly/pull/184))

## 0.8.1 (July 12, 2019)

BUG FIXES:

* resource/fastly_service_v1/condition: Support `PREFETCH` in `type` validation ([#171](https://github.com/terraform-providers/terraform-provider-fastly/issues/171))

## 0.8.0 (June 28, 2019)

NOTES:

* provider: This release includes only a Terraform SDK upgrade with compatibility for Terraform v0.12. The provider remains backwards compatible with Terraform v0.11 and there should not be any significant behavioural changes. ([#173](https://github.com/terraform-providers/terraform-provider-fastly/pull/173))

## 0.7.0 (June 25, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `splunk` argument ([#130](https://github.com/terraform-providers/terraform-provider-fastly/issues/130))
* resource/fastly_service_v1: Add `blobstoragelogging` argument ([#117](https://github.com/terraform-providers/terraform-provider-fastly/issues/117))
* resource/fastly_service_v1: Add `comment` argument ([#70](https://github.com/terraform-providers/terraform-provider-fastly/issues/70))
* resource/fastly_service_v1: Add `version_comment` argument ([#126](https://github.com/terraform-providers/terraform-provider-fastly/issues/126))
* resource/fastly_service_v1/backend: Add `override_host` argument ([#163](https://github.com/terraform-providers/terraform-provider-fastly/issues/163))
* resource/fastly_service_v1/condition: Add validation for `type` argument ([#148](https://github.com/terraform-providers/terraform-provider-fastly/issues/148))

NOTES:

* provider: Update `go-fastly` client to v1.0.0 ([#165](https://github.com/terraform-providers/terraform-provider-fastly/pull/165))

## 0.6.1 (May 29, 2019)

NOTES:

* provider: Switch codebase dependency management from `govendor` to Go modules ([#128](https://github.com/terraform-providers/terraform-provider-fastly/pull/128))
* provider: Update `go-fastly` client to v0.4.3 ([#154](https://github.com/terraform-providers/terraform-provider-fastly/pull/154))

## 0.6.0 (February 08, 2019)

ENHANCEMENTS:

* provider: Enable request/response logging ([#120](https://github.com/terraform-providers/terraform-provider-fastly/issues/120))
* resource/fastly_service_v1: Add `activate` argument ([#45](https://github.com/terraform-providers/terraform-provider-fastly/pull/45))

## 0.5.0 (January 08, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/papertrail: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/sumologic: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/gcslogging: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/bigquerylogging: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/syslog: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/logentries: Add `placement` argument ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))

BUG FIXES:

* resource/fastly_service_v1/snippet: Exclude dynamic snippets ([#107](https://github.com/terraform-providers/terraform-provider-fastly/pull/107))

## 0.4.0 (October 02, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `snippet` argument ([#93](https://github.com/terraform-providers/terraform-provider-fastly/pull/93))
* resource/fastly_service_v1: Add `director` argument ([#43](https://github.com/terraform-providers/terraform-provider-fastly/pull/43))
* resource/fastly_service_v1/bigquerylogging: Add `template` argument ([#90](https://github.com/terraform-providers/terraform-provider-fastly/pull/90))

BUG FIXES:

* resource/fastly_service_v1: Handle deletion of already deleted or never created resources ([#89](https://github.com/terraform-providers/terraform-provider-fastly/pull/89))

## 0.3.0 (August 02, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `bigquerylogging` argument ([#80](https://github.com/terraform-providers/terraform-provider-fastly/issues/80))

## 0.2.0 (June 04, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add `redundancy` argument ([64](https://github.com/terraform-providers/terraform-provider-fastly/pull/64))
* provider: Support for overriding base API URL ([68](https://github.com/terraform-providers/terraform-provider-fastly/pull/68))
* provider: Support for overriding user agent ([62](https://github.com/terraform-providers/terraform-provider-fastly/pull/62))

BUG FIXES:

* resource/fastly_service_v1/sumologic: Properly detect changes and update resource ([56](https://github.com/terraform-providers/terraform-provider-fastly/pull/56))

## 0.1.4 (January 16, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add StateFunc to hash secrets ([#63](https://github.com/terraform-providers/terraform-provider-fastly/issues/63))

## 0.1.3 (December 18, 2017)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `logentries` argument ([#24](https://github.com/terraform-providers/terraform-provider-fastly/issues/24))
* resource/fastly_service_v1: Add `syslog` argument ([#16](https://github.com/terraform-providers/terraform-provider-fastly/issues/16))

ENHANCEMENTS:

* resource/fastly_service_v1/syslog: Add `message_type` argument ([#30](https://github.com/terraform-providers/terraform-provider-fastly/issues/30))

## 0.1.2 (August 02, 2017)

ENHANCEMENTS:

* resource/fastly_service_v1/backend: Add `ssl_ca_cert` argument ([#11](https://github.com/terraform-providers/terraform-provider-fastly/issues/11))
* resource/fastly_service_v1/s3logging: Add `message_type` argument ([#14](https://github.com/terraform-providers/terraform-provider-fastly/issues/14))
* resource/fastly_service_v1/gcslogging: Add environment variable support for `secret_key` argument ([#15](https://github.com/terraform-providers/terraform-provider-fastly/issues/15))

BUG FIXES:

* resource/fastly_service_v1/s3logging: Update default value of `domain` argument ([#12](https://github.com/terraform-providers/terraform-provider-fastly/issues/12))

## 0.1.1 (June 21, 2017)

NOTES:

* provider: Bumping the provider version to get around provider caching issues - still same functionality

## 0.1.0 (June 20, 2017)

NOTES:

* provider: Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
