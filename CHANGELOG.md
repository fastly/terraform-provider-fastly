## 0.1.3 (Unreleased)

IMPROVEMENTS: 

* Add support for syslog logging [GH-16]

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
