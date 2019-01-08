## 0.6.0 (Unreleased)
## 0.5.0 (January 08, 2019)

IMPROVEMENTS:

* Add `placement` support to all logging providers ([#106](https://github.com/terraform-providers/terraform-provider-fastly/pull/106))

BUG FIXES:

* Skip dynamic VCL Snippets in order to prevent their deletion ([#107](https://github.com/terraform-providers/terraform-provider-fastly/pull/107))

## 0.4.0 (October 02, 2018)

IMPROVEMENTS:

* Add support for regular VCL Snippets ([#93](https://github.com/terraform-providers/terraform-provider-fastly/pull/93))
* Add `template` to Google BigQuery logging ([#90](https://github.com/terraform-providers/terraform-provider-fastly/pull/90))
* Add support for custom directors ([#43](https://github.com/terraform-providers/terraform-provider-fastly/pull/43))

BUG FIXES:

* Handle deletion of already deleted or never created resources ([#89](https://github.com/terraform-providers/terraform-provider-fastly/pull/89))

## 0.3.0 (August 02, 2018)

FEATURES:

* Add support for BigQuery logging ([#80](https://github.com/terraform-providers/terraform-provider-fastly/issues/80))

## 0.2.0 (June 04, 2018)

* s3_logging: Support for "redundancy" attribute ([64](https://github.com/terraform-providers/terraform-provider-fastly/pull/64))
* Support for overriding base API url ([68](https://github.com/terraform-providers/terraform-provider-fastly/pull/68))
* Support for overriding user agent ([62](https://github.com/terraform-providers/terraform-provider-fastly/pull/62))
* sumologic: Fixes sumologic implementation ([56](https://github.com/terraform-providers/terraform-provider-fastly/pull/56))

## 0.1.4 (January 16, 2018)

IMPROVEMENTS:

* s3_logging: Obfuscate secrets before persisting state file ([#63](https://github.com/terraform-providers/terraform-provider-fastly/issues/63))

## 0.1.3 (December 18, 2017)

IMPROVEMENTS: 

* Add support for Logentries logging ([#24](https://github.com/terraform-providers/terraform-provider-fastly/issues/24))
* Add support for format_version for Logentries logs ([#36](https://github.com/terraform-providers/terraform-provider-fastly/issues/36))
* Add support for syslog logging ([#16](https://github.com/terraform-providers/terraform-provider-fastly/issues/16))
* Add `message_type` support to syslog logging configuration ([#30](https://github.com/terraform-providers/terraform-provider-fastly/issues/30))

## 0.1.2 (August 02, 2017)

IMPROVEMENTS:

* backend: add `ssl_ca_cert` option ([#11](https://github.com/terraform-providers/terraform-provider-fastly/issues/11))
* s3_logging: Support message_type attribute ([#14](https://github.com/terraform-providers/terraform-provider-fastly/issues/14))
* gcs_logging: Optionally use env variable for credentials ([#15](https://github.com/terraform-providers/terraform-provider-fastly/issues/15))

BUG FIXES: 

* s3logging: default S3 domain to `s3.amazonaws.com` to match api default ([#12](https://github.com/terraform-providers/terraform-provider-fastly/issues/12))

## 0.1.1 (June 21, 2017)

NOTES:

Bumping the provider version to get around provider caching issues - still same functionality

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
