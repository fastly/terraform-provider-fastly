## 0.33.0 (Unreleased)

## 0.32.0 (June 17, 2021)

ENHANCEMENTS:

* Return 404 for non-existent service instead of a low-level nil entry error [#422](https://github.com/fastly/terraform-provider-fastly/pull/422)

BUG FIXES:

* Fix runtime panic in request-settings caused by incorrect type cast [#424](https://github.com/fastly/terraform-provider-fastly/pull/424)
* When `activate=true`, always read and clone from the active version [#423](https://github.com/fastly/terraform-provider-fastly/pull/423)

## 0.31.0 (June 14, 2021)

ENHANCEMENTS:

* Add support for ACL and extra redundancy options in S3 logging block [#417](https://github.com/fastly/terraform-provider-fastly/pull/417)
* Update default initial value for health check [#414](https://github.com/fastly/terraform-provider-fastly/pull/414)

BUG FIXES:

* Only set `cloned_version` after the version has been successfully validated [#418](https://github.com/fastly/terraform-provider-fastly/pull/418)

## 0.30.0 (May 12, 2021)

ENHANCEMENTS:

* Add director support for compute resource [#410](https://github.com/fastly/terraform-provider-fastly/pull/410)

## 0.29.1 (May 7, 2021)

BUG FIXES:

* Fix Header resource key names [#407](https://github.com/fastly/terraform-provider-fastly/pull/407)

## 0.29.0 (May 4, 2021)

ENHANCEMENTS:

* Add support for `file_max_bytes` configuration for Azure logging endpoint [#398](https://github.com/fastly/terraform-provider-fastly/pull/398)
* Support usage of IAM role in S3 and Kinesis logging endpoints [#403](https://github.com/fastly/terraform-provider-fastly/pull/403)
* Add support for `compression_codec` to logging file sink endpoints [#402](https://github.com/fastly/terraform-provider-fastly/pull/402)

DOCUMENTATION:

* Update debug mode instructions for Terraform 0.12.x [#405](https://github.com/fastly/terraform-provider-fastly/pull/405)

OTHER:

* Replace `master` with `main`. [#404](https://github.com/fastly/terraform-provider-fastly/pull/404)

## 0.28.2 (April 9, 2021)

BUG FIXES:

* Catch case where state from older version could be unexpected [#396](https://github.com/fastly/terraform-provider-fastly/pull/396)

## 0.28.1 (April 8, 2021)

BUG FIXES:

* Clone from `cloned_version` not `active_version` [#390](https://github.com/fastly/terraform-provider-fastly/pull/390)

## 0.28.0 (April 6, 2021)

ENHANCEMENTS:

* PATCH endpoint for TLS subscriptions [#370](https://github.com/fastly/terraform-provider-fastly/pull/370)
* Ensure passwords are marked as sensitive [#389](https://github.com/fastly/terraform-provider-fastly/pull/389)
* Add debug mode [#386](https://github.com/fastly/terraform-provider-fastly/pull/386)

BUG FIXES:

* Fix custom TLS configuration incorrectly omitting DNS records data [#392](https://github.com/fastly/terraform-provider-fastly/pull/392)
* Fix backend diff output incorrectly showing multiple resources being updated [#387](https://github.com/fastly/terraform-provider-fastly/pull/387)

DOCUMENTATION:

* Terraform 0.12+ no longer uses interpolation syntax for non-constant expressions [#384](https://github.com/fastly/terraform-provider-fastly/pull/384)

## 0.27.0 (March 16, 2021)

ENHANCEMENTS:

* Terraform Plugin SDK Upgrade [#379](https://github.com/fastly/terraform-provider-fastly/pull/379)
* Automate developer override for testing locally built provider [#382](https://github.com/fastly/terraform-provider-fastly/pull/382)

## 0.26.0 (March 5, 2021)

ENHANCEMENTS:

* Better sensitive value handling in Google pub/sub logging provider [#376](https://github.com/fastly/terraform-provider-fastly/pull/376)

BUG FIXES:

* Fix panic caused by incorrect type assert [#377](https://github.com/fastly/terraform-provider-fastly/pull/377)

## 0.25.0 (February 26, 2021)

ENHANCEMENTS:

* Add TLSCLientCert and TLSClientKey options for splunk logging ([#353](https://github.com/fastly/terraform-provider-fastly/pull/353))
* Add Dictionary to Compute service ([#361](https://github.com/fastly/terraform-provider-fastly/pull/361))
* Resources for Custom TLS and Platform TLS products ([#364](https://github.com/fastly/terraform-provider-fastly/pull/364))
* Managed TLS Subscriptions Resources ([#365](https://github.com/fastly/terraform-provider-fastly/pull/365))
* Ensure schema.Set uses custom SetDiff algorithm ([#366](https://github.com/fastly/terraform-provider-fastly/pull/366))
* Test speedup ([#371](https://github.com/fastly/terraform-provider-fastly/pull/371))
* Add service test sweeper ([#373](https://github.com/fastly/terraform-provider-fastly/pull/373))
* Add force_destroy flag to ACLs and Dicts to allow deleting non-empty lists ([#372](https://github.com/fastly/terraform-provider-fastly/pull/372))

## 0.24.0 (February 4, 2021)

ENHANCEMENTS:

* CI: check if docs need to be regenerated ([#362](https://github.com/fastly/terraform-provider-fastly/pull/362))
* Update go-fastly dependency to 3.0.0 ([#359](https://github.com/fastly/terraform-provider-fastly/pull/359))

BUG FIXES:

* Replace old doc generation process with new tfplugindocs tool ([#356](https://github.com/fastly/terraform-provider-fastly/pull/356))

## 0.23.0 (January 14, 2021)

ENHANCEMENT:

* Add support for kafka endpoints with sasl options ([#342](https://github.com/fastly/terraform-provider-fastly/pull/342))

## 0.22.0 (January 08, 2021)

ENHANCEMENT:

* Add Kinesis logging support ([#351](https://github.com/fastly/terraform-provider-fastly/pull/351))

## 0.21.3 (January 04, 2021)

NOTES:

* provider: Change version of go-fastly to v2.0.0 ([#341](https://github.com/fastly/terraform-provider-fastly/pull/341))

## 0.21.2 (December 16, 2020)

BUG FIXES:

* resource/fastly_service_*: Ensure we still refresh remote state when `activate` is set to `false` ([#345](https://github.com/fastly/terraform-provider-fastly/pull/345))

## 0.21.1 (October 15, 2020)

BUG FIXES:

* resource/fastly_service_waf_configuration: Guard `rule_exclusion` read to ensure API is only called if used. ([#330](https://github.com/fastly/terraform-provider-fastly/pull/330))

## 0.21.0 (October 14, 2020)

ENHANCEMENTS:

* resource/fastly_service_waf_configuration: Add `rule_exclusion` block which allows rules to be excluded from the WAF configuration. ([#328](https://github.com/fastly/terraform-provider-fastly/pull/328))

NOTES:

* provider: Change version of go-fastly to v2.0.0-alpha.1 ([#327](https://github.com/fastly/terraform-provider-fastly/pull/327))

## 0.20.4 (September 30, 2020)

NOTES:

* resource/fastly_service_acl_entries_v1: Change ACL documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
* resource/fastly_service_dictionary_items_v1: Change Dictionary documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
* resource/fastly_service_dynamic_snippet_content_v1: Change Dynamic Snippet documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
* resource/fastly_service_waf_configuration: Correctly mark `allowed_request_content_type_charset` as optional in documentation. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/322))

## 0.20.3 (September 23, 2020)

BUG FIXES:

* resource/fastly_service_v1/bigquerylogging: Ensure BigQuery logging `email`, `secret_key` fields are required and not optional. ([#319](https://github.com/fastly/terraform-provider-fastly/pull/319))

## 0.20.2 (September 22, 2020)

BUG FIXES:

* resource/fastly_service_v1: Improve performance of service read and delete logic. ([#311](https://github.com/fastly/terraform-provider-fastly/pull/311))
* resource/fastly_service_v1/logging_scalyr: Ensure `token` field is `sensitive` and thus hidden from plan. ([#310](https://github.com/fastly/terraform-provider-fastly/pull/310))

NOTES:

* resource/fastly_service_v1/s3logging: Document `server_side_encryption` and `server_side_encryption_kms_key_id` fields. ([#317](https://github.com/fastly/terraform-provider-fastly/pull/310))

## 0.20.1 (September 2, 2020)

BUG FIXES:

* resource/fastly_service_v1/backend: Ensure changes to backend fields result in updates instead of destroy and recreate. ([#304](https://github.com/fastly/terraform-provider-fastly/pull/304))
* resource/fastly_service_v1/logging_*: Fix logging acceptance tests by ensuring formatVersion in VCLLoggingAttributes is a *uint. ([#307](https://github.com/fastly/terraform-provider-fastly/pull/307))

NOTES:

* provider: Add a [CONTRIBUTING.md](https://github.com/fastly/terraform-provider-fastly/blob/main/CONTRIBUTING.md) containing contributing guidlines and documentation. ([#305](https://github.com/fastly/terraform-provider-fastly/pull/307))

## 0.20.0 (August 10, 2020)

FEATURES:

* **New Data Source:** `fastly_waf_rules` Use this data source to fetch Fastly WAF rules and pass to a `fastly_service_waf_configuration`. ([#291](https://github.com/fastly/terraform-provider-fastly/pull/291))
* **New Resource:** `fastly_service_waf_configuration` Provides a Web Application Firewall configuration and rules that can be applied to a service. ([#291](https://github.com/fastly/terraform-provider-fastly/pull/291))

ENHANCEMENTS:

* resource/fastly_service_v1/waf: Add `waf` block to enable and configure a Web Application Firewall on a service. ([#285](https://github.com/fastly/terraform-provider-fastly/pull/285))

## 0.19.3 (July 30, 2020)

NOTES:

* provider: Initial release to the [Terraform Registry](https://registry.terraform.io/)

## 0.19.2 (July 22, 2020)

NOTES:

* resource/fastly_service_compute: Fixes resource references in website documentation ([#296](https://github.com/fastly/terraform-provider-fastly/pull/296))

## 0.19.1 (July 22, 2020)

NOTES:

* resource/fastly_service_compute: Update website documentation for compute resource to include correct terminology ([#294](https://github.com/fastly/terraform-provider-fastly/pull/294))

## 0.19.0 (July 22, 2020)

FEATURES:

* **New Resource:** `fastly_service_compute` ([#281](https://github.com/fastly/terraform-provider-fastly/pull/281))

ENHANCEMENTS:

* resource/fastly_service_compute: Add support for all logging providers ([#285](https://github.com/fastly/terraform-provider-fastly/pull/285))
* resource/fastly_service_compute: Add support for importing compute services ([#286](https://github.com/fastly/terraform-provider-fastly/pull/286))
* resource/fastly_service_v1/ftp_logging: Add support for `message_type` field to FTP logging endpoint ([#288](https://github.com/fastly/terraform-provider-fastly/pull/288))

BUG FIXES:

* resource/fastly_service_v1/s3logging: Fix error check which was causing a runtime panic with s3logging ([#290](https://github.com/fastly/terraform-provider-fastly/pull/290))

NOTES:

* provider: Update `go-fastly` client to v1.16.2 ([#288](https://github.com/fastly/terraform-provider-fastly/pull/288))
* provider: Refactor documentation templating and compilation ([#283](https://github.com/fastly/terraform-provider-fastly/pull/283))

## 0.18.0 (July 01, 2020)

ENHANCEMENTS:

* resource/fastly_service_v1/logging_digitalocean: Add DigitalOcean Spaces logging support ([#276](https://github.com/fastly/terraform-provider-fastly/pull/276))
* resource/fastly_service_v1/logging_cloudfiles: Add Rackspace Cloud Files logging support ([#275](https://github.com/fastly/terraform-provider-fastly/pull/275))
* resource/fastly_service_v1/logging_openstack: Add OpenStack logging support ([#273](https://github.com/fastly/terraform-provider-fastly/pull/274))
* resource/fastly_service_v1/logging_logshuttle: Add Log Shuttle logging support ([#273](https://github.com/fastly/terraform-provider-fastly/pull/273))
* resource/fastly_service_v1/logging_honeycomb: Add Honeycomb logging support ([#272](https://github.com/fastly/terraform-provider-fastly/pull/272))
* resource/fastly_service_v1/logging_heroku: Add Heroku logging support ([#271](https://github.com/fastly/terraform-provider-fastly/pull/271))

NOTES:

* resource/fastly_service_v1/\*: "GZIP" -> "Gzip" ([#279](https://github.com/fastly/terraform-provider-fastly/pull/279))
* resource/fastly_service_v1/logging_sftp: Update SFTP logging to use `ValidateFunc` for validating the `message_type` field ([#278](https://github.com/fastly/terraform-provider-fastly/pull/278))
* resource/fastly_service_v1/gcslogging: Update GCS logging to use `ValidateFunc` for validating the `message_type` field ([#278](https://github.com/fastly/terraform-provider-fastly/pull/278))
* resource/fastly_service_v1/blobstoragelogging: Update Azure Blob Storage logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))
* resource/fastly_service_v1/logging_ftp: Update FTP logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))
* resource/fastly_service_v1/s3logging: Update S3 logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))

## 0.17.1 (June 24, 2020)

NOTES:

* resource/fastly_service_v1/\*: Migrates service resources to implement the `ServiceAttributeDefinition` `interface` ([#269](https://github.com/fastly/terraform-provider-fastly/pull/269))

## 0.17.0 (June 22, 2020)

ENHANCEMENTS:

* resource/fastly_service_v1/logging_googlepubsub: Add Google Cloud Pub/Sub logging support ([#258](https://github.com/fastly/terraform-provider-fastly/pull/258))
* resource/fastly_service_v1/logging_kafka: Add Kafka logging support ([#254](https://github.com/fastly/terraform-provider-fastly/pull/254))
* resource/fastly_service_v1/logging_scalyr: Add Scalyr logging support ([#252](https://github.com/fastly/terraform-provider-fastly/pull/252))
* resource/fastly_service_v1/s3logging: Add support for public key field ([#249](https://github.com/fastly/terraform-provider-fastly/pull/249))
* resource/fastly_service_v1/logging_newrelic: Add New Relic logging support ([#243](https://github.com/fastly/terraform-provider-fastly/pull/243))
* resource/fastly_service_v1/logging_datadog: Add Datadog logging support ([#242](https://github.com/fastly/terraform-provider-fastly/pull/242))
* resource/fastly_service_v1/logging_loggly: Add Loggly logging support ([#241](https://github.com/fastly/terraform-provider-fastly/pull/241))
* resource/fastly_service_v1/logging_sftp: Add SFTP logging support ([#236](https://github.com/fastly/terraform-provider-fastly/pull/236))
* resource/fastly_service_v1/logging_ftp: Add FTP logging support ([#235](https://github.com/fastly/terraform-provider-fastly/pull/235))
* resource/fastly_service_v1/logging_elasticsearch: Add Elasticsearch logging support ([#234](https://github.com/fastly/terraform-provider-fastly/pull/234))

NOTES:

* resource/fastly_service_v1/sftp: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `public_key` and `secret_key` fields ([#268](https://github.com/fastly/terraform-provider-fastly/pull/268))
* resource/fastly_service_v1/bigquerylogging: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `secret_key` field ([#268](https://github.com/fastly/terraform-provider-fastly/pull/268))
* resource/fastly_service_v1/httpslogging: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `tls_ca_cert`, `tls_client_cert` and `tls_client_key` fields ([#264](https://github.com/fastly/terraform-provider-fastly/pull/264))
* resource/fastly_service_v1/splunk: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `tls_ca_cert` field ([#264](https://github.com/fastly/terraform-provider-fastly/pull/264))
* resource/fastly_service_v1/\*: Migrate schemas to block separate block files ([#262](https://github.com/fastly/terraform-provider-fastly/pull/262))
* resource/fastly_service_v1/acl: Migrated to block file ([#253](https://github.com/fastly/terraform-provider-fastly/pull/253))
* provider: Update `go-fastly` client to v1.5.0 ([#248](https://github.com/fastly/terraform-provider-fastly/pull/248))

## 0.16.1 (June 03, 2020)

BUG FIXES:

* resource/fastly_service_v1/s3logging: Fix persistence of `server_side_encryption` and `server_side_encryption_kms_key_id` arguments ([#246](https://github.com/fastly/terraform-provider-fastly/pull/246))

## 0.16.0 (June 01, 2020)

ENHANCEMENTS:

* data-source/fastly_ip_ranges: Expose Fastly's IpV6 CIDR ranges via `ipv6_cidr_blocks` property ([#201](https://github.com/fastly/terraform-provider-fastly/pull/240))

NOTES:

* provider: Update `go-fastly` client to v1.14.1 ([#184](https://github.com/fastly/terraform-provider-fastly/pull/240))

## 0.15.0 (April 28, 2020)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `httpslogging` argument ([#222](https://github.com/fastly/terraform-provider-fastly/pull/222))
* resource/fastly_service_v1/splunk: Add `tls_hostname` and `tls_ca_cert` arguments ([#221](https://github.com/fastly/terraform-provider-fastly/pull/221))

NOTES:

* provider: Update `go-fastly` client to v1.10.0 ([#220](https://github.com/fastly/terraform-provider-fastly/pull/220))

## 0.14.0 (April 14, 2020)

FEATURES:

* **New Resource:** `fastly_user_v1` ([#214](https://github.com/fastly/terraform-provider-fastly/pull/214))

BUG FIXES:

* resource/fastly_service_v1/snippet: Fix support for `hash` snippet type ([#217](https://github.com/fastly/terraform-provider-fastly/pull/217))

NOTES:

* provider: Update `go` to v1.14.x ([#215](https://github.com/fastly/terraform-provider-fastly/pull/215))

## 0.13.0 (April 01, 2020)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add `server_side_encryption` and `server_side_encryption_kms_key_id` arguments ([#206](https://github.com/fastly/terraform-provider-fastly/pull/206))
* resource/fastly_service_v1/snippet: Support `hash` in `type` validation ([#211](https://github.com/fastly/terraform-provider-fastly/issues/211))
* resource/fastly_service_v1/dynamicsnippet: Support `hash` in `type` validation ([#211](https://github.com/fastly/terraform-provider-fastly/issues/211))

NOTES:

* provider: Update `go-fastly` client to v1.7.2 ([#213](https://github.com/fastly/terraform-provider-fastly/pull/213))

## 0.12.1 (January 23, 2020)

BUG FIXES:

* resource/fastly_service_v1: Allow a service to be created with a `default_ttl` of `0` ([#205](https://github.com/fastly/terraform-provider-fastly/pull/205))

## 0.12.0 (January 21, 2020)

ENHANCEMENTS:

* resource/fastly_service_v1/syslog: Add `tls_client_cert` and `tls_client_key` arguments ([#203](https://github.com/fastly/terraform-provider-fastly/pull/203))

## 0.11.1 (December 16, 2019)

BUG FIXES:

* data-source/fastly_ip_ranges: Use `go-fastly` client in order to fetch Fastly's assigned IP ranges ([#201](https://github.com/fastly/terraform-provider-fastly/pull/201))

## 0.11.0 (October 15, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1/dictionary: Add `write_only` argument ([#189](https://github.com/fastly/terraform-provider-fastly/pull/189))

NOTES:

* provider: The underlying Terraform codebase dependency for the provider SDK and acceptance testing framework has been migrated from `github.com/hashicorp/terraform` to `github.com/hashicorp/terraform-plugin-sdk`. They are functionality equivalent and this should only impact codebase development to switch imports. For more information see the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html). ([#191](https://github.com/fastly/terraform-provider-fastly/pull/191))
* provider: The actual Terraform version used by the provider will now be included in the `User-Agent` header for Terraform 0.12 and later. Terraform 0.11 and earlier will use `Terraform/0.11+compatible` as this information was not accessible in those versions. ([#182](https://github.com/fastly/terraform-provider-fastly/pull/182))

## 0.10.0 (October 02, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `cloned_version` argument ([#190](https://github.com/fastly/terraform-provider-fastly/pull/190))

## 0.9.0 (August 07, 2019)

FEATURES:

* **New Resource:** `fastly_service_acl_entries_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
* **New Resource:** `fastly_service_dictionary_items_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
* **New Resource:** `fastly_service_dynamic_snippet_content_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

ENHANCEMENTS:

* resource/fastly_service_v1: Add `acl` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
* resource/fastly_service_v1: Add `dictionary` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
* resource/fastly_service_v1: Add `dynamicsnippet` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

NOTES:

* provider: Update `go-fastly` client to v1.2.1 ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

## 0.8.1 (July 12, 2019)

BUG FIXES:

* resource/fastly_service_v1/condition: Support `PREFETCH` in `type` validation ([#171](https://github.com/fastly/terraform-provider-fastly/issues/171))

## 0.8.0 (June 28, 2019)

NOTES:

* provider: This release includes only a Terraform SDK upgrade with compatibility for Terraform v0.12. The provider remains backwards compatible with Terraform v0.11 and there should not be any significant behavioural changes. ([#173](https://github.com/fastly/terraform-provider-fastly/pull/173))

## 0.7.0 (June 25, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `splunk` argument ([#130](https://github.com/fastly/terraform-provider-fastly/issues/130))
* resource/fastly_service_v1: Add `blobstoragelogging` argument ([#117](https://github.com/fastly/terraform-provider-fastly/issues/117))
* resource/fastly_service_v1: Add `comment` argument ([#70](https://github.com/fastly/terraform-provider-fastly/issues/70))
* resource/fastly_service_v1: Add `version_comment` argument ([#126](https://github.com/fastly/terraform-provider-fastly/issues/126))
* resource/fastly_service_v1/backend: Add `override_host` argument ([#163](https://github.com/fastly/terraform-provider-fastly/issues/163))
* resource/fastly_service_v1/condition: Add validation for `type` argument ([#148](https://github.com/fastly/terraform-provider-fastly/issues/148))

NOTES:

* provider: Update `go-fastly` client to v1.0.0 ([#165](https://github.com/fastly/terraform-provider-fastly/pull/165))

## 0.6.1 (May 29, 2019)

NOTES:

* provider: Switch codebase dependency management from `govendor` to Go modules ([#128](https://github.com/fastly/terraform-provider-fastly/pull/128))
* provider: Update `go-fastly` client to v0.4.3 ([#154](https://github.com/fastly/terraform-provider-fastly/pull/154))

## 0.6.0 (February 08, 2019)

ENHANCEMENTS:

* provider: Enable request/response logging ([#120](https://github.com/fastly/terraform-provider-fastly/issues/120))
* resource/fastly_service_v1: Add `activate` argument ([#45](https://github.com/fastly/terraform-provider-fastly/pull/45))

## 0.5.0 (January 08, 2019)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/papertrail: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/sumologic: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/gcslogging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/bigquerylogging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/syslog: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
* resource/fastly_service_v1/logentries: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))

BUG FIXES:

* resource/fastly_service_v1/snippet: Exclude dynamic snippets ([#107](https://github.com/fastly/terraform-provider-fastly/pull/107))

## 0.4.0 (October 02, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `snippet` argument ([#93](https://github.com/fastly/terraform-provider-fastly/pull/93))
* resource/fastly_service_v1: Add `director` argument ([#43](https://github.com/fastly/terraform-provider-fastly/pull/43))
* resource/fastly_service_v1/bigquerylogging: Add `template` argument ([#90](https://github.com/fastly/terraform-provider-fastly/pull/90))

BUG FIXES:

* resource/fastly_service_v1: Handle deletion of already deleted or never created resources ([#89](https://github.com/fastly/terraform-provider-fastly/pull/89))

## 0.3.0 (August 02, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `bigquerylogging` argument ([#80](https://github.com/fastly/terraform-provider-fastly/issues/80))

## 0.2.0 (June 04, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add `redundancy` argument ([64](https://github.com/fastly/terraform-provider-fastly/pull/64))
* provider: Support for overriding base API URL ([68](https://github.com/fastly/terraform-provider-fastly/pull/68))
* provider: Support for overriding user agent ([62](https://github.com/fastly/terraform-provider-fastly/pull/62))

BUG FIXES:

* resource/fastly_service_v1/sumologic: Properly detect changes and update resource ([56](https://github.com/fastly/terraform-provider-fastly/pull/56))

## 0.1.4 (January 16, 2018)

ENHANCEMENTS:

* resource/fastly_service_v1/s3logging: Add StateFunc to hash secrets ([#63](https://github.com/fastly/terraform-provider-fastly/issues/63))

## 0.1.3 (December 18, 2017)

ENHANCEMENTS:

* resource/fastly_service_v1: Add `logentries` argument ([#24](https://github.com/fastly/terraform-provider-fastly/issues/24))
* resource/fastly_service_v1: Add `syslog` argument ([#16](https://github.com/fastly/terraform-provider-fastly/issues/16))

ENHANCEMENTS:

* resource/fastly_service_v1/syslog: Add `message_type` argument ([#30](https://github.com/fastly/terraform-provider-fastly/issues/30))

## 0.1.2 (August 02, 2017)

ENHANCEMENTS:

* resource/fastly_service_v1/backend: Add `ssl_ca_cert` argument ([#11](https://github.com/fastly/terraform-provider-fastly/issues/11))
* resource/fastly_service_v1/s3logging: Add `message_type` argument ([#14](https://github.com/fastly/terraform-provider-fastly/issues/14))
* resource/fastly_service_v1/gcslogging: Add environment variable support for `secret_key` argument ([#15](https://github.com/fastly/terraform-provider-fastly/issues/15))

BUG FIXES:

* resource/fastly_service_v1/s3logging: Update default value of `domain` argument ([#12](https://github.com/fastly/terraform-provider-fastly/issues/12))

## 0.1.1 (June 21, 2017)

NOTES:

* provider: Bumping the provider version to get around provider caching issues - still same functionality

## 0.1.0 (June 20, 2017)

NOTES:

* provider: Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
