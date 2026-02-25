## [UNRELEASED]

### BREAKING:

### ENHANCEMENTS:

### BUG FIXES:

- fix(ngwaf/rules): corrected the condition type assertion for nested single conditions in a group condition ([#1198](https://github.com/fastly/terraform-provider-fastly/pull/1198))

### DEPENDENCIES:

### DOCUMENTATION:

- docs(templates/guides): add a guide for adding a versionless domain to a service using a wildcard tls subscription ([#1194](https://github.com/fastly/terraform-provider-fastly/pull/1194))
- docs(templates/guides): add a guide for using versionless domains with a Certainly subscription to a new devlivery service ([#1195](https://github.com/fastly/terraform-provider-fastly/pull/1195))
- docs(templates/guides): add a guide for migrating delivery service classic domain to a versionless domain ([#1202](https://github.com/fastly/terraform-provider-fastly/pull/1202))
- docs(templates/guides): add a guide for linking versionless domains to a service when the domains are not managed in Terraform ([#1199](https://github.com/fastly/terraform-provider-fastly/pull/1199))
- docs(templates/guides): add a guide for migrating from the deprecated 'fastly_domain_v1' and 'fastly_domain_v1_service_link' resources and data sources  ([#1200](https://github.com/fastly/terraform-provider-fastly/pull/1200))
- docs(ngwaf/rules): updated list of supported values for the 'operator' field for NGWAF WAF rule conditions ([#1201](https://github.com/fastly/terraform-provider-fastly/pull/1201))

## 8.7.0 (February 20, 2026)

### ENHANCEMENTS:

- feat(product_enablement): Adding support for the `domain_inspector` feature to Compute services ([#1175](https://github.com/fastly/terraform-provider-fastly/pull/1175))
- feat(domain_management): Added import support for the `fastly_domain_service_link` resource and improved test coverage ([#1178](https://github.com/fastly/terraform-provider-fastly/pull/1178))
- feat(domains): Removed `_v1` suffixes from domain-related resources and data sources, leaving deprecated aliases in place. ([#1181](https://github.com/fastly/terraform-provider-fastly/pull/1181))
- feat(products/staging): Add a Data Source for Staging IP addresses ([#1186](https://github.com/fastly/terraform-provider-fastly/pull/1186))
- feat(ngwaf/rules): Added support for `multival` type conditions nested in `group_operator` blocks. ([#1189](https://github.com/fastly/terraform-provider-fastly/pull/1189))


### DEPENDENCIES:

- build(deps): `github.com/fastly/go-fastly/v12` from 12.1.0 to 12.1.1 ([#1177](https://github.com/fastly/terraform-provider-fastly/pull/1177))
- build(deps): `golang.org/x/net` from 0.48.0 to 0.49.0 ([#1177](https://github.com/fastly/terraform-provider-fastly/pull/1177))
- build(deps): `github.com/hashicorp/terraform-plugin-log` from 0.9.0 to 0.10.0 ([#1179](https://github.com/fastly/terraform-provider-fastly/pull/1179))
- build(deps): `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.38.1 to 2.38.2 ([#1180](https://github.com/fastly/terraform-provider-fastly/pull/1180))
- build(deps): `github.com/hashicorp/terraform-plugin-log` from 0.9.0 to 0.10.0 ([#1179](https://github.com/fastly/terraform-provider-fastly/pull/1179))
- build(go.mod): upgrade golang to 1.25.0 and make appropriate changes ([#1183](https://github.com/fastly/terraform-provider-fastly/pull/1183))
- build(deps): `github.com/fastly/go-fastly/v12` from 12.1.1 to 12.1.2 ([#1185](https://github.com/fastly/terraform-provider-fastly/pull/1185))
- build(deps): `golang.org/x/net` from 0.49.0 to 0.50.0 ([#1188](https://github.com/fastly/terraform-provider-fastly/pull/1188))
- build(deps): `github.com/fastly/go-fastly/v12` from 12.1.2 to 13.0.0 ([#1190](https://github.com/fastly/terraform-provider-fastly/pull/1190))


## 8.6.0 (December 17, 2025)

### ENHANCEMENTS:

- feat(provider): redact `Fastly-Key` from `TF_LOG=DEBUG` output ([#1167](https://github.com/fastly/terraform-provider-fastly/pull/1167))
- feat(provider): log response body for HTTP error responses (≥400) in `TF_LOG=DEBUG` output ([#1170](https://github.com/fastly/terraform-provider-fastly/pull/1170))

### BUG FIXES:

- fix(service/backend): corrected a drift issue caused by the `keepalive_time` attribute ([#1156](https://github.com/fastly/terraform-provider-fastly/pull/1156))
- fix(request_setting): preserve optional bool fields (`force_miss`, `force_ssl`, `bypass_busy_wait`, `timer_support`) during updates and add acceptance test coverage ([#1165](https://github.com/fastly/terraform-provider-fastly/pull/1165))

### DEPENDENCIES:

- build(deps): `actions/checkout` from 5 to 6 ([#1159](https://github.com/fastly/terraform-provider-fastly/pull/1159))
- build(deps): `golang.org/x/net` from 0.47.0 to 0.48.0 ([#1166](https://github.com/fastly/terraform-provider-fastly/pull/1166))
- build(deps): `actions/checkout` from 5 to 6 ([#1159](https://github.com/fastly/terraform-provider-fastly/pull/1159))

### DOCUMENTATION:

- docs(ngwaf/rules): provided examples of client_identifiers types for rate limit rules ([#1169](https://github.com/fastly/terraform-provider-fastly/pull/1169))
- docs(logging/bigquery): updates the details of the `account_name` field to match API requirements ([#1171](https://github.com/fastly/terraform-provider-fastly/pull/1171))

## 8.5.0 (November 20, 2025)

### ENHANCEMENTS:

- feat(compute_acl_entries): add CIDR validation ([#1136](https://github.com/fastly/terraform-provider-fastly/pull/1136))

### BUG FIXES:

- fix(header): preserve optional bool field `ignore_if_set` during updates and add acceptance test coverage ([#1142](https://github.com/fastly/terraform-provider-fastly/pull/1142))
- fix(block_fastly_service_logging_logentries_test): fix tests for logentries to account for API behavior ([#1143](https://github.com/fastly/terraform-provider-fastly/pull/1143))
- fix(image_optimizer_default_settings): preserve optional bool fields (`allow_video`, `webp`, `upscale`) during updates and add acceptance test coverage ([#1145](https://github.com/fastly/terraform-provider-fastly/pull/1145))
- fix(logging_kafka): preserve optional bool fields (`use_tls`, `parse_log_keyvals`) during updates and add acceptance test coverage ([#1147](https://github.com/fastly/terraform-provider-fastly/pull/1147))
- fix(product_enablement): ensure `ddos_protection` mode updates are applied ([#1149](https://github.com/fastly/terraform-provider-fastly/pull/1149))

### DEPENDENCIES:

- build(deps): `golangci/golangci-lint-action` from 8 to 9 ([#1144](https://github.com/fastly/terraform-provider-fastly/pull/1144))
- build(deps): `golang.org/x/net` from 0.46.0 to 0.47.0 ([#1150](https://github.com/fastly/terraform-provider-fastly/pull/1150))
- build(deps): `golangci/golangci-lint-action` from 8 to 9 ([#1144](https://github.com/fastly/terraform-provider-fastly/pull/1144))
- build(deps): `golang.org/x/crypto` from 0.44.0 to 0.45.0 ([#1154](https://github.com/fastly/terraform-provider-fastly/pull/1154))

### DOCUMENTATION:

- docs(ngwaf/rules): added signal exclusion rule type documentation ([#1140](https://github.com/fastly/terraform-provider-fastly/pull/1140))

## 8.4.0 (November 4, 2025)

### ENHANCEMENTS:

- feat(ngwaf/lists): added support for NGWAF Lists to data sources ([#1124](https://github.com/fastly/terraform-provider-fastly/pull/1124))
- feat(ngwaf/rules): added support for NGWAF Rules to data sources ([#1124](https://github.com/fastly/terraform-provider-fastly/pull/1124))
- feat(ngwaf/signals): added support for NGWAF Signals to data sources ([#1124](https://github.com/fastly/terraform-provider-fastly/pull/1124))

### BUG FIXES:

- fix(logging_https): ensure `response_condition` is applied during updates and add acceptance test coverage ([#1130](https://github.com/fastly/terraform-provider-fastly/pull/1130))
- fix(backend): preserve optional bool fields (`use_ssl`, `ssl_check_cert`, `prefer_ipv6`, `auto_loadbalance`) during updates and add acceptance test coverage ([#1133](https://github.com/fastly/terraform-provider-fastly/pull/1133))
- fix(domains_v1/service_link): corrected a behavior where new service links created were not referring to 'domain_id' values correctly ([#1132](https://github.com/fastly/terraform-provider-fastly/pull/1132))
- fix(logging/compute): corrected drift behavior for some compute logging endpoints where the value of 'period' was not being retained correctly ([#1134](https://github.com/fastly/terraform-provider-fastly/pull/1134))
- fix(tls/subscriptions): corrects 'common_name' validation for update operations ([#1135](https://github.com/fastly/terraform-provider-fastly/pull/1135))

### DOCUMENTATION:

- docs(ngwaf/rules): imroved usage examples of various NGWAF Rule patterns ([#1128](https://github.com/fastly/terraform-provider-fastly/pull/1128))

## 8.3.2 (October 16, 2025)

### DOCUMENTATION:

- fix: correct release with non main tag

## 8.3.1 (October 15, 2025)

### BUG FIXES:

- fix(logging/https): corrected a bug where users that had a HTTPS logging block would encounter 'gzip_level' API errors after upgrading to the v8.1.0 provider or later ([#1118](https://github.com/fastly/terraform-provider-fastly/pull/1118))

### DEPENDENCIES:

- build(deps): `stefanzweifel/git-auto-commit-action` from 6 to 7 ([#1120](https://github.com/fastly/terraform-provider-fastly/pull/1120))
- build(deps): `golang.org/x/net` from 0.44.0 to 0.46.0 ([#1119](https://github.com/fastly/terraform-provider-fastly/pull/1119))

## 8.3.0 (September 30, 2025)

### ENHANCEMENTS:

- feat(logging/https): add support for Period HTTPS logging endpoint ([#1097](https://github.com/fastly/terraform-provider-fastly/pull/1097))
- feat(product_enablement): Add enable/disable support for API Discovery ([#1111](https://github.com/fastly/terraform-provider-fastly/pull/1111))
- feat(domainsv1/data source): add support for the v1 domains data source ([#1112](https://github.com/fastly/terraform-provider-fastly/pull/1112))
- feat(service): 'domain' blocks are now optional ([#1113](https://github.com/fastly/terraform-provider-fastly/pull/1113))
- feat(domain_service_link): add support for domain service links ([#1110](https://github.com/fastly/terraform-provider-fastly/pull/1110))

## 8.2.0 (September 24, 2025)

### ENHANCEMENTS:

- feat(ngwaf/rules): add support for multival type conditions ([#1100](https://github.com/fastly/terraform-provider-fastly/pull/1100))

### BUG FIXES:

- fix(ngwaf/thresholds): Make duration optional and set default when missing ([#1107](https://github.com/fastly/terraform-provider-fastly/pull/1107))
- fix(ngwaf/alert_datadog_integration): Expand allowed Datadog key length ([#1107](https://github.com/fastly/terraform-provider-fastly/pull/1107))
- fix(ngwaf/alerts): Ensure that FASTLY_TF_DISPLAY_SENSITIVE_FIELDS is respected ([#1106](https://github.com/fastly/terraform-provider-fastly/pull/1106))

### DEPENDENCIES:

- build(deps): `github.com/fastly/go-fastly/v11` from 11.3.1 to 12.0.0 ([#1104](https://github.com/fastly/terraform-provider-fastly/pull/1104))
- build(deps): `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.37.0 to 2.38.1 ([#1108](https://github.com/fastly/terraform-provider-fastly/pull/1108))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.3.1 to 12.0.0 ([#1104](https://github.com/fastly/terraform-provider-fastly/pull/1104))

## 8.1.0 (September 17, 2025)

### ENHANCEMENTS:

- feat(ngwaf/workspace): fix basic usage example
- feat(ngwaf/workspace): add default_redirect_url to workspaces. ([#1068](https://github.com/fastly/terraform-provider-fastly/pull/1068))
- feat(compute acls): add support for compute ACLs ([#1031](https://github.com/fastly/terraform-provider-fastly/pull/1031))
- refactor(resource_fastly_domain_v1) to use domainmanagement imports and types ([#1074](https://github.com/fastly/terraform-provider-fastly/pull/1074))
- feat(ngwaf/rules): add support for deception and allow_interactive actions ([#1077](https://github.com/fastly/terraform-provider-fastly/pull/1077))
- feat(compute/healthcheck): add support for healthchecks for Compute services ([#1079](https://github.com/fastly/terraform-provider-fastly/pull/1079))
- feat(logging): add support for compression to HTTPS logging endpoint ([#1086](https://github.com/fastly/terraform-provider-fastly/pull/1086))
- feat(compute/logging_newrelicotlp): add support for New Relic OLTP logging for Compute services ([#1095](https://github.com/fastly/terraform-provider-fastly/pull/1095))

### BUG FIXES:

- fix(ngwaf/rules): removes the rate limit block from the account level rules schema. ([#1065](https://github.com/fastly/terraform-provider-fastly/pull/1065))
- fix(service_vcl/logging_gcs): resolves an issue where project_id was not being updated for GCS logging_gcs. ([#1073](https://github.com/fastly/terraform-provider-fastly/pull/1073))
- fix(service_vcl/requestsetting): removed a incorrect default value for the xff attribute. ([#1078](https://github.com/fastly/terraform-provider-fastly/pull/1078))
- fix(ngwaf/workspace): corrected zero values being set from workspace imports when attack thresholds are left as default ([#1103](https://github.com/fastly/terraform-provider-fastly/pull/1103))

### DEPENDENCIES:

- build(deps): `github.com/fastly/go-fastly/v11` from 11.1.1 to 11.2.0 ([#1067](https://github.com/fastly/terraform-provider-fastly/pull/1067))
- build(deps): `golang.org/x/net` from 0.42.0 to 0.43.0 ([#1072](https://github.com/fastly/terraform-provider-fastly/pull/1072))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.1.1 to 11.2.0 ([#1067](https://github.com/fastly/terraform-provider-fastly/pull/1067))
- build(deps): `actions/checkout` from 4 to 5 ([#1071](https://github.com/fastly/terraform-provider-fastly/pull/1071))
- build(deps): `github.com/stretchr/testify` from 1.10.0 to 1.11.0 ([#1081](https://github.com/fastly/terraform-provider-fastly/pull/1081))
- build(deps): `github.com/stretchr/testify` from 1.11.0 to 1.11.1 ([#1085](https://github.com/fastly/terraform-provider-fastly/pull/1085))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.3.0 to 11.3.1 ([#1087](https://github.com/fastly/terraform-provider-fastly/pull/1087))
- build(deps): `actions/setup-go` from 5 to 6 ([#1092](https://github.com/fastly/terraform-provider-fastly/pull/1092))
- build(deps): `actions/github-script` from 7 to 8 ([#1092](https://github.com/fastly/terraform-provider-fastly/pull/1092))
- build(deps): `golang.org/x/net` from 0.43.0 to 0.44.0 ([#1099](https://github.com/fastly/terraform-provider-fastly/pull/1099))

### DOCUMENTATION:

- docs(ngwaf/rules): add rule action types documentation ([#1069](https://github.com/fastly/terraform-provider-fastly/pull/1069))
- docs(ngwaf/rules): updated rule action type references ([#1098](https://github.com/fastly/terraform-provider-fastly/pull/1098))

## 8.0.0 (August 5, 2025)

### ENHANCEMENTS:

- feat(ngwaf): Add support for Next-Gen WAF (many PRs).
- doc(guides): Add guide for Fastly Object Storage. ([#1024](https://github.com/fastly/terraform-provider-fastly/pull/1024))
- doc(resources): Improve DDoS protection configuration documentation. ([#1029](https://https://github.com/fastly/terraform-provider-fastly/pull/1029))

### BUG FIXES:

- fix(snippets): delete dynamic snippet contents when the resource is deleted if `manage_snippets` is `true` ([#1021](https://github.com/fastly/terraform-provider-fastly/pull/1021))
- fix(examples): Replace http-me.glitch.me with http-me.fastly.dev ([#1026](https://github.com/fastly/terraform-provider-fastly/pull/1026))
- fix(product_enablement/ngwaf): Allow traffic_ramp to be set to zero ([#1057](https://github.com/fastly/terraform-provider-fastly/pull/1057))

### DEPENDENCIES:

- feat(deps): Upgrade to go-fastly version 11. ([#1028](https://github.com/fastly/terraform-provider-fastly/pull/1028))
- build(deps): `golang.org/x/net` from 0.41.0 to 0.42.0 ([#1030](https://github.com/fastly/terraform-provider-fastly/pull/1030))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.0.0 to 11.1.0 ([#1059](https://github.com/fastly/terraform-provider-fastly/pull/1059))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.1.0 to 11.1.1 ([#1060](https://github.com/fastly/terraform-provider-fastly/pull/1060))

## 8.0.0-beta (August 5, 2025)

### ENHANCEMENTS:

- feat(ngwaf): Add support for Next-Gen WAF (many PRs).
- doc(guides): Add guide for Fastly Object Storage. ([#1024](https://github.com/fastly/terraform-provider-fastly/pull/1024))
- doc(resources): Improve DDoS protection configuration documentation. ([#1029](https://https://github.com/fastly/terraform-provider-fastly/pull/1029))

### BUG FIXES:

- fix(snippets): delete dynamic snippet contents when the resource is deleted if `manage_snippets` is `true`. ([#1021](https://github.com/fastly/terraform-provider-fastly/pull/1021))
- fix(examples): Replace http-me.glitch.me with http-me.fastly.dev ([#1026](https://github.com/fastly/terraform-provider-fastly/pull/1026))
- fix(product_enablement/ngwaf): Allow traffic_ramp to be set to zero ([#1057](https://github.com/fastly/terraform-provider-fastly/pull/1057))

### DEPENDENCIES:

- feat(deps): Upgrade to go-fastly version 11. ([#1028](https://github.com/fastly/terraform-provider-fastly/pull/1028))
- build(deps): `golang.org/x/net` from 0.41.0 to 0.42.0 ([#1030](https://github.com/fastly/terraform-provider-fastly/pull/1030))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.0.0 to 11.1.0 ([#1059](https://github.com/fastly/terraform-provider-fastly/pull/1059))
- build(deps): `github.com/fastly/go-fastly/v11` from 11.1.0 to 11.1.1 ([#1060](https://github.com/fastly/terraform-provider-fastly/pull/1060))

## 7.1.0 (June 20, 2025)

### ENHANCEMENTS:

- feat(domains/v1): add `description` field ([#1002](https://github.com/fastly/terraform-provider-fastly/pull/1002))
- feat(backend): add support for 'prefer IPv6' attribute ([#1003](https://github.com/fastly/terraform-provider-fastly/pull/1003))
- feat(logging): add support for 'processing region' attribute ([#1011](https://github.com/fastly/terraform-provider-fastly/pull/1011))

### BUG FIXES:

- fix(block_fastly_service_settings): fix detection of falsey value for stale_if_error field ([#1003](https://github.com/fastly/terraform-provider-fastly/pull/1009))
- fix(backend): set default 'prefer IPv6' attribute differently for Delivery and Compute services ([#1010](https://github.com/fastly/terraform-provider-fastly/pull/1010))

### DEPENDENCIES:

- build(deps): `github.com/fastly/go-fastly/v10` from 10.2.0 to 10.3.0 ([#1004](https://github.com/fastly/terraform-provider-fastly/pull/1004))
- build(deps): `github.com/cloudflare/circl` from 1.6.0 to 1.6.1 ([#1006](https://github.com/fastly/terraform-provider-fastly/pull/1006))
- build(deps): `github.com/fastly/go-fastly/v10` from 10.2.0 to 10.3.0 ([#1004](https://github.com/fastly/terraform-provider-fastly/pull/1004))
- build(deps): `golang.org/x/net` from 0.40.0 to 0.41.0 ([#1005](https://github.com/fastly/terraform-provider-fastly/pull/1005))
- build(deps): `github.com/fastly/go-fastly/v10` from 10.3.0 to 10.4.0 ([#1012](https://github.com/fastly/terraform-provider-fastly/pull/1012))

## 7.0.0 (May 21, 2025)

### BREAKING:

- feat(fastly): remove deprecated `geo_headers` field from fastly service ([#992](https://github.com/fastly/terraform-provider-fastly/pull/992))

### ENHANCEMENTS:

- feat(config): add an environment variable allowing users to override the sensitive attribute ([#985](https://github.com/fastly/terraform-provider-fastly/pull/985))

### DEPENDENCIES:

- build(deps): `golang.org/x/net` from 0.39.0 to 0.40.0 ([#990](https://github.com/fastly/terraform-provider-fastly/pull/990))
- build(deps): `github.com/fastly/go-fastly/v10` from 10.0.1 to 10.1.0 ([#996](https://github.com/fastly/terraform-provider-fastly/pull/996))
- build(deps): `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.36.1 to 2.37.0 ([#998](https://github.com/fastly/terraform-provider-fastly/pull/998))
- build(deps): `github.com/fastly/go-fastly/v10` from 10.1.0 to 10.2.0 ([#1000](https://github.com/fastly/terraform-provider-fastly/pull/1000))
- build(go.mod): upgrade golang to 1.24.0 and make appropriate changes ([#993](https://github.com/fastly/terraform-provider-fastly/pull/993))

## 6.1.0 (April 17, 2025)

### ENHANCEMENTS:

- feat(logging): restore support for `placement` attribute ([#965](https://github.com/fastly/terraform-provider-fastly/pull/965))

### BUG FIXES:

- fix(block_fastly_service_snippet_test.go): breaking changes introduced by `go-fastly` v10.0.1 ([#981](https://github.com/fastly/terraform-provider-fastly/pull/981))
- fix(block_fastly_service_dynamicsnippet.go): breaking changes introduced by `go-fastly` v10.0.1 ([#981](https://github.com/fastly/terraform-provider-fastly/pull/981))
- fix(data_source_vcl_snippets.go): breaking changes introduced by `go-fastly` v10.0.1 ([#982](https://github.com/fastly/terraform-provider-fastly/pull/982))

### DEPENDENCIES:

- build(deps): `golang.org/x/net` from 0.37.0 to 0.38.0 ([#966](https://github.com/fastly/terraform-provider-fastly/pull/966))
- build(deps): `go-fastly` from 9.14.0 to 10.0.0 ([#970](https://github.com/fastly/terraform-provider-fastly/pull/970))
- build(deps): `golang.org/x/net` from 0.38.0 to 0.39.0 ([#974](https://github.com/fastly/terraform-provider-fastly/pull/974))
- build(deps): `actions/create-github-app-token` from 1 to 2 ([#973](https://github.com/fastly/terraform-provider-fastly/pull/973))
- build(deps): `actions/github-script` from 6 to 7 ([#976](https://github.com/fastly/terraform-provider-fastly/pull/976))
- build(deps): `github.com/fastly/go-fastly/v10` from 10.0.0 to 10.0.1 ([#977](https://github.com/fastly/terraform-provider-fastly/pull/977))

## 6.0.1 (March 25, 2025)

### BUG FIXES:

- fix(dictionary): add error check preventing deletion of write only dictionaries without force destroy [#959](https://github.com/fastly/terraform-provider-fastly/pull/959)
- fix(product_enablement): first check if state exists before accessing it [#961](https://github.com/fastly/terraform-provider-fastly/pull/961)

### DEPENDENCIES:

- build(deps): `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.34.0 to 2.36.1 ([#960](https://github.com/fastly/terraform-provider-fastly/pull/960))

## 6.0.0 (March 20, 2025)

### BREAKING:

- breaking(waf): support for the Fastly WAF (legacy, not Next-Gen WAF)
  product has been removed. The product passed its End-of-Life date
  quite some time ago, and it is no longer in use by customers
  [#936](https://github.com/fastly/terraform-provider-fastly/pull/936)

- breaking(logging): the 'placement' attribute in the logging
  endpoints has been changed to ignore any value provided by the user;
  it was only used in combination with the Fastly WAF, which is no
  longer supported
  [#936](https://github.com/fastly/terraform-provider-fastly/pull/936)

### ENHANCEMENTS:

- feat(product_enablement): add DDoS protection product enablement/configuration ([#954](https://github.com/fastly/terraform-provider-fastly/pull/954))
- feat(product_enablement): add Next-Gen WAF product enablement/configuration ([#956](https://github.com/fastly/terraform-provider-fastly/pull/956))
- feat(object_storage_access_key): add object storage access keys configuration ([#955](https://github.com/fastly/terraform-provider-fastly/pull/955))

### DEPENDENCIES:

- build(go.mod) upgrade to go 1.23.0
- build(deps): `github.com/hashicorp/go-cty` from 1.4.1-0.20200414143053-d3edf31b6320 to 1.4.1 ([#946](https://github.com/fastly/terraform-provider-fastly/pull/946))
- build(deps): `github.com/fastly/go-fastly/v9` from 9.13.1 to 9.14.0 ([#947](https://github.com/fastly/terraform-provider-fastly/pull/947))
- build(deps): `golang.org/x/net` from 0.35.0 to 0.37.0 ([#948](https://github.com/fastly/terraform-provider-fastly/pull/948))
- build(deps): `github.com/deckarep/golang-set/v2` from 2.7.0 to 2.8.0 ([#953](https://github.com/fastly/terraform-provider-fastly/pull/953))
- build(deps): `github.com/hashicorp/go-cty` from 1.4.1 to 1.5.0 ([#952](https://github.com/fastly/terraform-provider-fastly/pull/952))

## 5.17.0 (March 10, 2025)

### ENHANCEMENTS:

- feat(staging): add support for 'staging' of service versions ([#933](https://github.com/fastly/terraform-provider-fastly/pull/933))

### BUG FIXES:

- fix(fastly_vcl_service): always 'validate' services after applying changes ([#942](https://github.com/fastly/terraform-provider-fastly/pull/942))

### DEPENDENCIES:

- build(deps): `github.com/hashicorp/terraform-plugin-docs` from 0.19.4 to 0.21.0 ([#937](https://github.com/fastly/terraform-provider-fastly/pull/937))
- build(deps): `github.com/google/go-cmp` from 0.6.0 to 0.7.0 ([#932](https://github.com/fastly/terraform-provider-fastly/pull/932))
- build(deps): `golang.org/x/net` from 0.34.0 to 0.35.0 ([#921](https://github.com/fastly/terraform-provider-fastly/pull/921))
- build(deps): `github.com/fastly/go-fastly/v9` from 9.13.0 to 9.13.1 ([#927](https://github.com/fastly/terraform-provider-fastly/pull/927))
- build(deps): `github.com/bflad/tfproviderlint` from 0.30.0 to 0.31.0 ([#928](https://github.com/fastly/terraform-provider-fastly/pull/928))

## 5.16.0 (January 31, 2025)

ENHANCEMENTS:

- feat(domains): add support for v1 functionality [#917](https://github.com/fastly/terraform-provider-fastly/pull/917)
- feat(dashboard): add support for Observability custom dashboards [#905](https://github.com/fastly/terraform-provider-fastly/pull/905)
- feat(alerts): append 'Managed by Terraform' to descriptions [#914](https://github.com/fastly/terraform-provider-fastly/pull/914)

BUG FIXES:

- fix(fastly_package_hash): unnecessary `source_code_hash` conflict [#909](https://github.com/fastly/terraform-provider-fastly/pull/909)

DEPENDENCIES:

- build(deps): bump github.com/stretchr/testify from 1.8.4 to 1.10.0 [#901](https://github.com/fastly/terraform-provider-fastly/pull/901)

DOCUMENTATION:

- docs: correct links in Product Enablement documentation [#907](https://github.com/fastly/terraform-provider-fastly/pull/907)

## 5.15.0 (November 12, 2024)

ENHANCEMENTS:

- Support for Grafana Cloud Logs. [#895](https://github.com/fastly/terraform-provider-fastly/pull/895)
- feat(product_enablement): Add support for Log Explorer & Insights product. [#896](https://github.com/fastly/terraform-provider-fastly/pull/896)

BUG FIXES:

- fix(tls_mutual_authentication): Ensure that 'enforced' property does not revert to default during changes. [#890](https://github.com/fastly/terraform-provider-fastly/pull/890)
- breaking(product_enablement): Remove support for NGWAF product. [#893](https://github.com/fastly/terraform-provider-fastly/pull/893)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.29.0 to 0.30.0 [#889](https://github.com/fastly/terraform-provider-fastly/pull/889)
- build(deps): Update to go-fastly 9.12.0. [#894](https://github.com/fastly/terraform-provider-fastly/pull/894)
- build(deps): bump golang.org/x/net from 0.30.0 to 0.31.0 [#897](https://github.com/fastly/terraform-provider-fastly/pull/897)

## 5.14.0 (October 3, 2024)

ENHANCEMENTS:

- feat(product_enablement): Add support for NGWAF product. [#886](https://github.com/fastly/terraform-provider-fastly/pull/886)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.28.0 to 0.29.0 [#882](https://github.com/fastly/terraform-provider-fastly/pull/882)
- build(deps): bump github.com/fastly/go-fastly/v9 from 9.9.0 to 9.10.0 [#883](https://github.com/fastly/terraform-provider-fastly/pull/883)
- build(deps): Update to go-fastly 9.11.0. [#886](https://github.com/fastly/terraform-provider-fastly/pull/886)

## 5.13.0 (August 26, 2024)

ENHANCEMENTS:

- feat(product_enablement) - Add support for Fastly Bot Management. [#880](https://github.com/fastly/terraform-provider-fastly/pull/880)

DEPENDENCIES:

- build(deps): bump github.com/fastly/go-fastly/v9 from 9.8.0 to 9.9.0 (part of #880)

## 5.12.0 (August 15, 2024)

ENHANCEMENTS:

- feat(tls_activation): support mutual_authentication_id in domain activation at creation [#875](https://github.com/fastly/terraform-provider-fastly/pull/875)

BUG FIXES:

- Clarify Image Optimizer shield requirements [#874](https://github.com/fastly/terraform-provider-fastly/pull/874)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.26.0 to 0.27.0 [#872](https://github.com/fastly/terraform-provider-fastly/pull/872)
- build(deps): bump golang.org/x/net from 0.27.0 to 0.28.0 [#877](https://github.com/fastly/terraform-provider-fastly/pull/877)
- build(deps): bump github.com/fastly/go-fastly/v9 from 9.7.0 to 9.8.0 [#878](https://github.com/fastly/terraform-provider-fastly/pull/878)

## 5.11.0 (July 2, 2024)

BUG FIXES:

- fix(alerts): make service_id attribute optional [#852](https://github.com/fastly/terraform-provider-fastly/pull/852)
- fix handling of config store entries when store has been deleted [#864](https://github.com/fastly/terraform-provider-fastly/pull/864)
- fix(tls_activation): update mutual_authentication_id [#866](https://github.com/fastly/terraform-provider-fastly/pull/866)
- fix(alerts): ensure that alert creation works properly when 'ignore_below' is not specified [#869](https://github.com/fastly/terraform-provider-fastly/pull/869)

DEPENDENCIES:

- build(deps): bump github.com/bflad/tfproviderlint from 0.29.0 to 0.30.0 [#850](https://github.com/fastly/terraform-provider-fastly/pull/850)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 [#853](https://github.com/fastly/terraform-provider-fastly/pull/853)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs [#857](https://github.com/fastly/terraform-provider-fastly/pull/857)
- build(deps): bump golang.org/x/net from 0.25.0 to 0.26.0 [#859](https://github.com/fastly/terraform-provider-fastly/pull/859)
- build(deps): bump github.com/fastly/go-fastly/v9 from 9.4.0 to 9.7.0 [#867](https://github.com/fastly/terraform-provider-fastly/pull/867)

## 5.10.0 (May 16, 2024)

ENHANCEMENTS:

- feat(image_optimizer_default_settings): add Image Optimizer default settings API [#832](https://github.com/fastly/terraform-provider-fastly/pull/832)

## 5.9.0 (May 9, 2024)

ENHANCEMENTS:

- feat(fastly_integration): implement resource and documentation [#844](https://github.com/fastly/terraform-provider-fastly/pull/844)
- feat(fastly_alerts): add percentage alerts [#845](https://github.com/fastly/terraform-provider-fastly/pull/845)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.19.0 to 0.24.0 [#842](https://github.com/fastly/terraform-provider-fastly/pull/842)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs [#841](https://github.com/fastly/terraform-provider-fastly/pull/841)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs [#843](https://github.com/fastly/terraform-provider-fastly/pull/843)
- build(deps): bump golang.org/x/net from 0.24.0 to 0.25.0 [#846](https://github.com/fastly/terraform-provider-fastly/pull/846)

## 5.8.0 (April 18, 2024)

BUG FIXES:

- fix(fastly_vcl_service): ensure backend names are unique [#836](https://github.com/fastly/terraform-provider-fastly/pull/836)

ENHANCEMENTS:

- feat: vcl_snippets [#835](https://github.com/fastly/terraform-provider-fastly/pull/835)
- Add support for specifying location when creating KV stores [#834](https://github.com/fastly/terraform-provider-fastly/pull/834)

## 5.7.3 (April 16, 2024)

BUG FIXES:

- Fix stats alerts, add all service aggregate alerts [#831](https://github.com/fastly/terraform-provider-fastly/pull/831)

## 5.7.2 (April 15, 2024)

ENHANCEMENTS:

- fix: update the default timeout value for healthcheck consistency with Fastly App (UI) [#827](https://github.com/fastly/terraform-provider-fastly/pull/827)

BUG FIXES:

- fix(rate_limiter): persist uri_dictionary_name to state [#828](https://github.com/fastly/terraform-provider-fastly/pull/828)
- fix(tls_subscription): ensure configuration_id is current value (not initial) [#824](https://github.com/fastly/terraform-provider-fastly/pull/824)
- fix(tls_mutual_authentication): update activation after mtls creation [#829](https://github.com/fastly/terraform-provider-fastly/pull/829)

DOCUMENTATION:

- Update Certainly Documentation to Remove Beta Label [#826](https://github.com/fastly/terraform-provider-fastly/pull/826)

## 5.7.1 (March 15, 2024)

ENHANCEMENTS:

- feat(scalyr): add project_id [#822](https://github.com/fastly/terraform-provider-fastly/pull/822)

DEPENDENCIES:

- chore: avoid extra string interpolation [#820](https://github.com/fastly/terraform-provider-fastly/pull/820)

## 5.7.0 (February 29, 2024)

BUG FIXES:

- remove: mTLS from state if API returns 404 [#794](https://github.com/fastly/terraform-provider-fastly/pull/794)
- fix(docs): YAML Frontmatter formatting in product_enablement.md [#788](https://github.com/fastly/terraform-provider-fastly/pull/788)
- fix(request_settings): allow unsetting of action [#814](https://github.com/fastly/terraform-provider-fastly/pull/814)

ENHANCEMENTS:

- ci: add golangci-lint action [#777](https://github.com/fastly/terraform-provider-fastly/pull/777)
- feat(logging_newrelicotlp): add new logging block [#786](https://github.com/fastly/terraform-provider-fastly/pull/786)
- ci: slash command to trigger tests for forked PRs [#785](https://github.com/fastly/terraform-provider-fastly/pull/785)
- ci: fix ok-to-test [#796](https://github.com/fastly/terraform-provider-fastly/pull/796)
- refactor(all): support go-fastly v9 [#808](https://github.com/fastly/terraform-provider-fastly/pull/808)
- feat(fastly_alert): implement resource and documentation [#810](https://github.com/fastly/terraform-provider-fastly/pull/810)

DEPENDENCIES:

- build(deps): bump hashicorp/setup-terraform from 2 to 3 [#776](https://github.com/fastly/terraform-provider-fastly/pull/776)
- build(deps): bump github.com/fastly/go-fastly/v8 from 8.6.2 to 8.6.4 [#779](https://github.com/fastly/terraform-provider-fastly/pull/779)
- build(deps): bump google.golang.org/grpc from 1.57.0 to 1.57.1 [#780](https://github.com/fastly/terraform-provider-fastly/pull/780)
- build(deps): bump actions/setup-go from 4 to 5 [#789](https://github.com/fastly/terraform-provider-fastly/pull/789)
- build(deps): bump golang.org/x/crypto from 0.14.0 to 0.17.0 [#793](https://github.com/fastly/terraform-provider-fastly/pull/793)
- build(deps): bump golang.org/x/net from 0.17.0 to 0.19.0 [#784](https://github.com/fastly/terraform-provider-fastly/pull/784)

## 5.6.0 (October 18, 2023)

BUG FIXES:

- fix(product_enablement): avoid accidentally disabling products on update [#763](https://github.com/fastly/terraform-provider-fastly/pull/763)

ENHANCEMENTS:

- refactor(product_enablement): make Read() logic consistent with other resource types [#773](https://github.com/fastly/terraform-provider-fastly/pull/773)

DEPENDENCIES:

- build(deps): bump github.com/fastly/go-fastly/v8 from 8.6.1 to 8.6.2 [#765](https://github.com/fastly/terraform-provider-fastly/pull/765)
- build(deps): bump golang.org/x/net from 0.15.0 to 0.17.0 [#771](https://github.com/fastly/terraform-provider-fastly/pull/771)
- build(deps): bump github.com/google/go-cmp from 0.5.9 to 0.6.0 [#770](https://github.com/fastly/terraform-provider-fastly/pull/770)

DOCUMENTATION:

- docs: product enablement [#762](https://github.com/fastly/terraform-provider-fastly/pull/762)
- doc: rename Compute@Edge to Compute [#769](https://github.com/fastly/terraform-provider-fastly/pull/769)

## 5.5.0 (September 19, 2023)

ENHANCEMENTS:

- feat(backend): support share_key attribute [#747](https://github.com/fastly/terraform-provider-fastly/pull/747)
- test(interface): add more resources [#746](https://github.com/fastly/terraform-provider-fastly/pull/746)
- test(interface): add more fastly_service_vcl attributes/blocks [#756](https://github.com/fastly/terraform-provider-fastly/pull/756)
- test(interface): add rate_limiter resource [#759](https://github.com/fastly/terraform-provider-fastly/pull/759)

BUG FIXES:

- fix: use paginator to fetch all ACL entries [#758](https://github.com/fastly/terraform-provider-fastly/pull/758)

DEPENDENCIES:

- build(deps): bump actions/checkout from 3 to 4 [#744](https://github.com/fastly/terraform-provider-fastly/pull/744)
- build(deps): bump github.com/fastly/go-fastly/v8 from 8.5.9 to 8.6.1 [#745](https://github.com/fastly/terraform-provider-fastly/pull/745)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.28.0 to 2.29.0 [#752](https://github.com/fastly/terraform-provider-fastly/pull/752)
- build(deps): bump goreleaser/goreleaser-action from 4 to 5 [#754](https://github.com/fastly/terraform-provider-fastly/pull/754)
- build(deps): bump golang.org/x/net from 0.14.0 to 0.15.0 [#753](https://github.com/fastly/terraform-provider-fastly/pull/753)

## 5.4.0 (September 1, 2023)

ENHANCEMENTS:

- feat(fastly_secretstore): implement resource and documentation [#707](https://github.com/fastly/terraform-provider-fastly/pull/707)
- ci: validate interface not broken [#735](https://github.com/fastly/terraform-provider-fastly/pull/735)

BUG FIXES:

- fix(product_enablement): add additional error message filter [#740](https://github.com/fastly/terraform-provider-fastly/pull/740)

DEPENDENCIES:

- build: update all dependencies [#739](https://github.com/fastly/terraform-provider-fastly/pull/739)
- build(deps): bump golang.org/x/net from 0.12.0 to 0.14.0 [#734](https://github.com/fastly/terraform-provider-fastly/pull/734)

## 5.3.1 (August 9, 2023)

ENHANCEMENTS:

- feat(package): make package optional [#733](https://github.com/fastly/terraform-provider-fastly/pull/733)

BUG FIXES:

- revert(backend): revert removal of `error_threshold` attribute [#731](https://github.com/fastly/terraform-provider-fastly/pull/731)

## 5.3.0 (August 4, 2023)

ENHANCEMENTS:

- feat: create fastly_configstores data source [#729](https://github.com/fastly/terraform-provider-fastly/pull/729)
- feat: create fastly_kvstores data source [#730](https://github.com/fastly/terraform-provider-fastly/pull/730)

BUG FIXES:

- fix(request_settings): don't send empty string for request_condition [#722](https://github.com/fastly/terraform-provider-fastly/pull/722)
- fix(backend): remove redundant error_threshold attribute [#731](https://github.com/fastly/terraform-provider-fastly/pull/731)

DEPENDENCIES:

- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 [#723](https://github.com/fastly/terraform-provider-fastly/pull/723)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs [#726](https://github.com/fastly/terraform-provider-fastly/pull/726)
- build(deps): bump golang.org/x/net from 0.11.0 to 0.12.0 [#725](https://github.com/fastly/terraform-provider-fastly/pull/725)
- build(deps): bump github.com/fastly/go-fastly/v8 from 8.5.4 to 8.5.7 [#727](https://github.com/fastly/terraform-provider-fastly/pull/727)

## 5.2.2 (June 29, 2023)

BUG FIXES:

- fix(stores): remove store from state if not found remotely [#719](https://github.com/fastly/terraform-provider-fastly/pull/719)

DEPENDENCIES:

- build(deps): bump go-fastly to latest 8.5.4 [#720](https://github.com/fastly/terraform-provider-fastly/pull/720)

## 5.2.1 (June 23, 2023)

DEPENDENCIES:

- build(deps): update go-fastly to latest 8.5.2 release [#717](https://github.com/fastly/terraform-provider-fastly/pull/717)

## 5.2.0 (June 22, 2023)

ENHANCEMENTS:

- feat: add file_max_bytes attribute to logging_s3 resource [#711](https://github.com/fastly/terraform-provider-fastly/pull/711)

BUG FIXES:

- fix(rate_limiter): add rate limter ID to delete call [#714](https://github.com/fastly/terraform-provider-fastly/pull/714)
- fix(rate_limiter): lookup new ID before actioning a deletion [#715](https://github.com/fastly/terraform-provider-fastly/pull/715)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.10.0 to 0.11.0 [#712](https://github.com/fastly/terraform-provider-fastly/pull/712)

## 5.1.0 (June 13, 2023)

ENHANCEMENTS:

- feat(kv_store): support KV Store [#691](https://github.com/fastly/terraform-provider-fastly/pull/691)
- feat(mutual_authentication): implement mTLS resource [#702](https://github.com/fastly/terraform-provider-fastly/pull/702)
- feat(config_store): implement config store resource [#705](https://github.com/fastly/terraform-provider-fastly/pull/705)

BUG FIXES:

- fix(rate_limiter): fix multiple runtime panics [#706](https://github.com/fastly/terraform-provider-fastly/pull/706)

DEPENDENCIES:

- build(deps): bump github.com/stretchr/testify from 1.8.2 to 1.8.3 [#700](https://github.com/fastly/terraform-provider-fastly/pull/700)
- build(deps): bump github.com/stretchr/testify from 1.8.3 to 1.8.4 [#704](https://github.com/fastly/terraform-provider-fastly/pull/704)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.14.1 to 0.15.0 [#709](https://github.com/fastly/terraform-provider-fastly/pull/709)

## 5.0.0 (May 22, 2023)

BREAKING:

There was a long-standing issue with how Terraform reacted to the `package.tar.gz` file that the CLI produces. Effectively, hashing the package was inconsistent and caused Terraform to think the code had changed even when it hadn't.

To resolve the issue the Package API now returns a new metadata property (`files_hash`) that calculates the hash from a sorted list of the files within the package.

This PR updates the Terraform provider to use this new property instead of the original `hashsum` metadata property and exposes a new `fastly_package_hash` data source that will generate the appropriate value for the `source_code_hash` attribute.

Although the public interface has not changed, the underlying implementation changes have meant customers will no longer be able to use the previous approach of using `filesha512` to generate a hash from their package file. So we must consider this PR a breaking change.

This does require a slight change to a customer's process, which prior to this release looked like this...

```tf
source_code_hash = filesha512("package.tar.gz")
```

As of this release, we recommend the use of the `fastly_package_hash` data source...

```tf
data "fastly_package_hash" "example" {
  filename = "./path/to/package.tar.gz"
}

resource "fastly_service_compute" "example" {
  # ...

  package {
    filename         = "./path/to/package.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
}
```

- breaking(compute): fix package hash bug [#698](https://github.com/fastly/terraform-provider-fastly/pull/698)

## 4.3.3 (May 12, 2023)

BUG FIXES:

- fix(gcs): `project_id` should be optional [#693](https://github.com/fastly/terraform-provider-fastly/pull/693)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.9.0 to 0.10.0 [#692](https://github.com/fastly/terraform-provider-fastly/pull/692)

## 4.3.2 (May 2, 2023)

BUG FIXES:

- fix(product_enablement): avoid unexpected diff [#689](https://github.com/fastly/terraform-provider-fastly/pull/689)

## 4.3.1 (April 29, 2023)

ENHANCEMENTS:

- Bump go-fastly to new v8 major release to add project-id in GCS for logging [#685](https://github.com/fastly/terraform-provider-fastly/pull/685)

BUG FIXES:

- fix(product_enablement): error message check was too specific [#687](https://github.com/fastly/terraform-provider-fastly/pull/687)

## 4.3.0 (April 19, 2023)

ENHANCEMENTS:

- feat(data_source): new dictionaries data source [#682](https://github.com/fastly/terraform-provider-fastly/pull/682)

DEPENDENCIES:

- build(deps): bump golang.org/x/net from 0.8.0 to 0.9.0 [#680](https://github.com/fastly/terraform-provider-fastly/pull/680)
- build(deps): bump github.com/bflad/tfproviderlint from 0.28.1 to 0.29.0 [#681](https://github.com/fastly/terraform-provider-fastly/pull/681)

## 4.2.0 (April 1, 2023)

ENHANCEMENTS:

- feat(ratelimiter): implement Rate Limiter API [#678](https://github.com/fastly/terraform-provider-fastly/pull/678)

## 4.1.2 (March 29, 2023)

BUG FIXES:

- fix(fastly_service_vcl): validate snippet names [#673](https://github.com/fastly/terraform-provider-fastly/pull/673)
- fix(fastly_service_vcl): don't call http3 endpoint if already enabled [#675](https://github.com/fastly/terraform-provider-fastly/pull/675)

DEPENDENCIES:

- Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.26.0 to 2.26.1 [#671](https://github.com/fastly/terraform-provider-fastly/pull/671)

## 4.1.1 (March 27, 2023)

BUG FIXES:

- fix(tls/subscriptions): tls configuration id should always be passed [#670](https://github.com/fastly/terraform-provider-fastly/pull/670)

DOCUMENTATION:

- docs(tls/subscriptions): clarify default tls config [commit](https://github.com/fastly/terraform-provider-fastly/commit/e0c80d39060082271f382965d21d341e1affaeb0)

DEPENDENCIES:

- build(dependencies): bump github.com/hashicorp/terraform-plugin-sdk/v2 [#667](https://github.com/fastly/terraform-provider-fastly/pull/667)
- Bump actions/setup-go from 3 to 4 [#664](https://github.com/fastly/terraform-provider-fastly/pull/664)
- Bump github.com/fastly/go-fastly/v7 from 7.4.0 to 7.5.0 [#665](https://github.com/fastly/terraform-provider-fastly/pull/665)

## 4.1.0 (March 16, 2023)

ENHANCEMENTS:

- feat(fastly_service_compute): support new `content` attribute [#661](https://github.com/fastly/terraform-provider-fastly/pull/661)

DOCUMENTATION:

- docs(dictionary): add import note [#662](https://github.com/fastly/terraform-provider-fastly/pull/662)

DEPENDENCIES:

- Bump github.com/hashicorp/terraform-plugin-docs from 0.13.0 to 0.14.1 [#656](https://github.com/fastly/terraform-provider-fastly/pull/656)
- Bump golang.org/x/net from 0.0.0-20211112202133-69e39bad7dc2 to 0.8.0 [#655](https://github.com/fastly/terraform-provider-fastly/pull/655)
- Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.24.0 to 2.25 [#642](https://github.com/fastly/terraform-provider-fastly/pull/642)
- Bump github.com/stretchr/testify from 1.8.1 to 1.8.2 [#647](https://github.com/fastly/terraform-provider-fastly/pull/647)

## 4.0.0 (March 9, 2023)

BREAKING:

Only one minor breaking interface change has been made: the removal of the `auto_loadbalance` attribute from the `backend` block, which is still supported for the `fastly_service_vcl` resource but was never actually supported under the `fastly_service_compute` resource.

- fix(backend): remove `auto_loadbalance` from compute service [#657](https://github.com/fastly/terraform-provider-fastly/pull/657)

BUG FIXES:

- fix: add missing format attribute when updating [#659](https://github.com/fastly/terraform-provider-fastly/pull/659)

ENHANCEMENTS:

- Enable the declaration of the keepalive_time backend attribute [#658](https://github.com/fastly/terraform-provider-fastly/pull/658)

## 3.2.0 (March 2, 2023)

ENHANCEMENTS:

- Allow certainly as a certificate authority [#648](https://github.com/fastly/terraform-provider-fastly/pull/648)

BUG FIXES:

- fix(product_enablement): improve error handling for user scenarios without self-enablement [#651](https://github.com/fastly/terraform-provider-fastly/pull/651)

DOCUMENTATION:

- docs: tls subscriptions [#649](https://github.com/fastly/terraform-provider-fastly/pull/649)

## 3.1.0 (February 21, 2023)

ENHANCEMENTS:

- feat(http3): implementing the HTTP3 API [#640](https://github.com/fastly/terraform-provider-fastly/pull/640)
- feat(product_enablement): implement product enablement APIs [#641](https://github.com/fastly/terraform-provider-fastly/pull/641)

## 3.0.4 (January 9, 2023)

BUG FIXES:

- fix: force refresh when service version is reverted outside of Terraform [#630](https://github.com/fastly/terraform-provider-fastly/pull/630)

DEPENDENCIES:

- Bump goreleaser/goreleaser-action from 3 to 4 [#626](https://github.com/fastly/terraform-provider-fastly/pull/626)

## 3.0.3 (December 7, 2022)

BUG FIXES:

- Prevent SSL related fields from being sent empty to the Fastly API [#622](https://github.com/fastly/terraform-provider-fastly/pull/622)

DOCUMENTATION:

- docs: remove 'alpha' notice from custom health check http headers feature [#623](https://github.com/fastly/terraform-provider-fastly/pull/623)

## 3.0.2 (November 23, 2022)

BUG FIXES:

- Fix logging endpoints to not send empty placement value [#620](https://github.com/fastly/terraform-provider-fastly/pull/620)

## 3.0.1 (November 22, 2022)

BUG FIXES:

- Backends send empty string to API [#618](https://github.com/fastly/terraform-provider-fastly/pull/618)

## 3.0.0 (November 16, 2022)

The major v7 release of the go-fastly API client resulted in substantial changes to the internals of the Fastly Terraform provider, and so we felt it was safer to release a new major version.

Additionally, the long deprecated `ssl_hostname` backend attribute has now officially been removed from the provider (refer to the documentation for `ssl_cert_hostname` and `ssl_sni_hostname`).

There has also been many bug fixes as part of the integration with the latest go-fastly release.

BREAKING:

- Bump go-fastly to new v7 major release [#614](https://github.com/fastly/terraform-provider-fastly/pull/614)

ENHANCEMENTS:

- feat: dependabot workflow automation for updating dependency [#604](https://github.com/fastly/terraform-provider-fastly/pull/604)
- Add google account name to all gcp logging endpoints [#603](https://github.com/fastly/terraform-provider-fastly/pull/603)

BUG FIXES:

- fix incorrect update reference [#599](https://github.com/fastly/terraform-provider-fastly/pull/599)

DEPENDENCIES:

- Bump actions/checkout from 2 to 3 [#605](https://github.com/fastly/terraform-provider-fastly/pull/605)
- Bump goreleaser/goreleaser-action from 2 to 3 [#606](https://github.com/fastly/terraform-provider-fastly/pull/606)
- Bump github.com/bflad/tfproviderlint from 0.27.1 to 0.28.1 [#611](https://github.com/fastly/terraform-provider-fastly/pull/611)
- Bump github.com/stretchr/testify from 1.7.0 to 1.8.1 [#610](https://github.com/fastly/terraform-provider-fastly/pull/610)
- Bump github.com/google/go-cmp from 0.5.6 to 0.5.9 [#608](https://github.com/fastly/terraform-provider-fastly/pull/608)
- Bump actions/setup-go from 2 to 3 [#607](https://github.com/fastly/terraform-provider-fastly/pull/607)
- Bump github.com/hashicorp/terraform-plugin-docs from 0.5.0 to 0.13.0 [#612](https://github.com/fastly/terraform-provider-fastly/pull/612)
- Bump actions/cache from 2 to 3 [#616](https://github.com/fastly/terraform-provider-fastly/pull/616)

## 2.4.0 (October 13, 2022)

ENHANCEMENTS:

- Support health check headers [#598](https://github.com/fastly/terraform-provider-fastly/pull/598)
- Code base refactor with tfproviderlintx [#596](https://github.com/fastly/terraform-provider-fastly/pull/596)

## 2.3.3 (October 3, 2022)

ENHANCEMENTS:

- Reduce unnecessary API calls [#593](https://github.com/fastly/terraform-provider-fastly/pull/593)

## 2.3.2 (September 20, 2022)

ENHANCEMENTS:

- Support for additional S3 storage classes [#589](https://github.com/fastly/terraform-provider-fastly/pull/589)

DOCUMENTATION:

- Remove inconsistent 'warning' message [#591](https://github.com/fastly/terraform-provider-fastly/pull/591)

## 2.3.1 (September 12, 2022)

BUG FIXES:

- Bump dependencies to fix Dependabot vulnerability [#586](https://github.com/fastly/terraform-provider-fastly/pull/586)

## 2.3.0 (September 5, 2022)

ENHANCEMENTS:

- Allow services to be reused [#578](https://github.com/fastly/terraform-provider-fastly/pull/578)
- Static analysis refactoring [#581](https://github.com/fastly/terraform-provider-fastly/pull/581)
- Revive linter refactoring [#584](https://github.com/fastly/terraform-provider-fastly/pull/584)

DOCUMENTATION:

- Document updating `fastly_tls_certificate` [#582](https://github.com/fastly/terraform-provider-fastly/pull/582)

## 2.2.1 (July 21, 2022)

BUG FIXES:

- Fix Splunk `token` attribute to be required [#579](https://github.com/fastly/terraform-provider-fastly/pull/579)

## 2.2.0 (July 5, 2022)

ENHANCEMENTS:

- Data Source: fastly_services [#575](https://github.com/fastly/terraform-provider-fastly/pull/575)

## 2.1.0 (June 27, 2022)

ENHANCEMENTS:

- Support Service Authorizations [#572](https://github.com/fastly/terraform-provider-fastly/pull/572)

BUG FIXES:

- Fix integration tests [#573](https://github.com/fastly/terraform-provider-fastly/pull/573)

## 2.0.0 (May 10, 2022)

BUG FIXES:

- Remove unsupported features (Healthchecks, Directors and VCL settings) from `fastly_service_compute` [#569](https://github.com/fastly/terraform-provider-fastly/pull/569)

## 1.1.4 (April 28, 2022)

ENHANCEMENTS:

- Avoid unnecessary API calls for `director` block. [#567](https://github.com/fastly/terraform-provider-fastly/pull/567)

BUG FIXES:

- Fix `fastly_tls_configuration` pagination logic [#565](https://github.com/fastly/terraform-provider-fastly/pull/565)

## 1.1.3 (April 25, 2022)

ENHANCEMENTS:

- The `backend` block is no longer required within the `fastly_service_compute` resource [#563](https://github.com/fastly/terraform-provider-fastly/pull/563)

DOCUMENTATION:

- Clarify key/cert update flow [#561](https://github.com/fastly/terraform-provider-fastly/pull/562)
- Typo in `fastly_tls_certificate` resource [#560](https://github.com/fastly/terraform-provider-fastly/pull/560)
- Typo in `manage_entries` attribute [#559](https://github.com/fastly/terraform-provider-fastly/pull/559)

## 1.1.2 (March 4, 2022)

BUG FIXES:

- Add Terraform provider version to User-Agent [#553](https://github.com/fastly/terraform-provider-fastly/pull/553)

## 1.1.1 (March 3, 2022)

DOCUMENTATION:

- Add 1.0.0 Migration Guide to Documentation [#551](https://github.com/fastly/terraform-provider-fastly/pull/551)

## 1.1.0 (February 23, 2022)

ENHANCEMENTS:

- Add `fastly_datacenters` data resource [#540](https://github.com/fastly/terraform-provider-fastly/pull/540)

BUG FIXES:

- Support removing backends from a `director` [#547](https://github.com/fastly/terraform-provider-fastly/pull/547)

DOCUMENTATION:

- Fix `fastly-s3` hyperlinks [#542](https://github.com/fastly/terraform-provider-fastly/pull/542)

## 1.0.0 (February 8, 2022)

ENHANCEMENTS:

- Changes for v1.0.0 [#534](https://github.com/fastly/terraform-provider-fastly/pull/534)

BUG FIXES:

- Fix the example usage in docs/index.md [#533](https://github.com/fastly/terraform-provider-fastly/pull/533)
- Support Terraform CLI 1.1.4 [#536](https://github.com/fastly/terraform-provider-fastly/pull/536)

## 0.41.0 (January 21, 2022)

ENHANCEMENTS:

- Add activate attribute support for `fastly_service_waf_configuration` resource [#530](https://github.com/fastly/terraform-provider-fastly/pull/530)
- Support updating TLS subscriptions in pending state [#528](https://github.com/fastly/terraform-provider-fastly/pull/528)

DOCUMENTATION:

- Revamp fastly_tls_subscription examples [#527](https://github.com/fastly/terraform-provider-fastly/pull/527)

## 0.40.0 (January 13, 2022)

ENHANCEMENTS:

- Bump go-fastly to v6.0.0 [#525](https://github.com/fastly/terraform-provider-fastly/pull/525)
- Add new `modsec_rule_ids` filter to `fastly_waf_rules` [#521](https://github.com/fastly/terraform-provider-fastly/pull/521)
- Force creation of new condition if type changed [#518](https://github.com/fastly/terraform-provider-fastly/pull/518)

DOCUMENTATION:

- Simplify example in `fastly_tls_subscription` [#516](https://github.com/fastly/terraform-provider-fastly/pull/516)
- Update VCL snippet type description [#519](https://github.com/fastly/terraform-provider-fastly/pull/519)

## 0.39.0 (December 8, 2021)

ENHANCEMENTS:

- Bump go-fastly to v5.1.3 [#511](https://github.com/fastly/terraform-provider-fastly/pull/511)
- Expose `certificate_id` read-only attribute from `fastly_tls_subscription` [#506](https://github.com/fastly/terraform-provider-fastly/pull/506)
- Dynamically generate provider version [#500](https://github.com/fastly/terraform-provider-fastly/pull/500)

BUG FIXES:

- Fix constants formatting [#504](https://github.com/fastly/terraform-provider-fastly/pull/504)
- Remove `-i` flag from `go test` [#503](https://github.com/fastly/terraform-provider-fastly/pull/503)

DOCUMENTATION:

- Consistent description for attributes using constants [#502](https://github.com/fastly/terraform-provider-fastly/pull/502)
- Update `RELEASE.md` [#499](https://github.com/fastly/terraform-provider-fastly/pull/499)

## 0.38.0 (November 4, 2021)

BUG FIXES:

- Do not send 0 `subnet` value unless explicitly set [#496](https://github.com/fastly/terraform-provider-fastly/pull/496)

DOCUMENTATION:

- Utilise `codefile` and `tffile` functions from tfplugindocs [#497](https://github.com/fastly/terraform-provider-fastly/pull/497)

## 0.37.0 (November 1, 2021)

BUG FIXES:

- Ignore 404 on GetPackage when importing wasm service [#487](https://github.com/fastly/terraform-provider-fastly/pull/487)
- Properly set `IdleConnTimeout` to prevent resource exhaustion on tests [#491](https://github.com/fastly/terraform-provider-fastly/pull/491)

ENHANCEMENTS:

- Remove TLS subscriptions that 404 from state [#479](https://github.com/fastly/terraform-provider-fastly/pull/479)
- Override `Transport` to enable keepalive and add new `force_http2` provider option [#485](https://github.com/fastly/terraform-provider-fastly/pull/485)
- Rename GNUmakefile to Makefile [#483](https://github.com/fastly/terraform-provider-fastly/pull/483)
- Only update service `name` and `comment` if `activate` is true [#481](https://github.com/fastly/terraform-provider-fastly/pull/481)
- Add `use_tls` attribute for Splunk logging [#482](https://github.com/fastly/terraform-provider-fastly/pull/482)

DOCUMENTATION:

- Convert `index.md` to template to inject provider version [#492](https://github.com/fastly/terraform-provider-fastly/pull/492)

## 0.36.0 (September 27, 2021)

BUG FIXES:

- Bump go-fastly to v5 to fix API client bugs [#477](https://github.com/fastly/terraform-provider-fastly/pull/477)
- Update `terraform-json` dependency so test suite runs successfully with Terraform v1 [#474](https://github.com/fastly/terraform-provider-fastly/pull/474)

ENHANCEMENTS:

- Add support for `stale-if-error` [#475](https://github.com/fastly/terraform-provider-fastly/pull/475)

DOCUMENTATION:

- Clarify edge private dictionary usage [#472](https://github.com/fastly/terraform-provider-fastly/pull/472)
- Correct ACL typos [#473](https://github.com/fastly/terraform-provider-fastly/pull/473)

## 0.35.0 (September 15, 2021)

ENHANCEMENTS:

- Make `backend` block optional [#457](https://github.com/fastly/terraform-provider-fastly/pull/457)
- Audit `sensitive` attributes [#458](https://github.com/fastly/terraform-provider-fastly/pull/458)
- Tests should not error when no backends defined (now considered as warning) [#462](https://github.com/fastly/terraform-provider-fastly/pull/462)
- Refactor service attribute handlers into CRUD-style functions [#463](https://github.com/fastly/terraform-provider-fastly/pull/463)
- Change to accept multi-pem blocks [#469](https://github.com/fastly/terraform-provider-fastly/pull/469)
- Bump go-fastly version [#467](https://github.com/fastly/terraform-provider-fastly/pull/467)

BUG FIXES:

- Fix `fastly_service_waf_configuration` not updating `rule` attributes correctly [#464](https://github.com/fastly/terraform-provider-fastly/pull/464)
- Correctly update `version_comment` [#466](https://github.com/fastly/terraform-provider-fastly/pull/466)

DEPRECATED:

- Deprecate `geo_headers` attribute [#456](https://github.com/fastly/terraform-provider-fastly/pull/456)

## 0.34.0 (August 9, 2021)

ENHANCEMENTS:

- Avoid unnecessary state refresh when importing (and enable service version selection) [#448](https://github.com/fastly/terraform-provider-fastly/pull/448)

BUG FIXES:

- Fix TLS Subscription updates not triggering update to managed DNS Challenges [#453](https://github.com/fastly/terraform-provider-fastly/pull/453)

## 0.33.0 (July 16, 2021)

ENHANCEMENTS:

- Upgrade to Go 1.16 to allow `darwin/arm64` builds [#447](https://github.com/fastly/terraform-provider-fastly/pull/447)
- Replace `ActivateVCL` call with `Main` field on `CreateVCL` [#446](https://github.com/fastly/terraform-provider-fastly/pull/446)
- Add limitations for `write_only` dictionaries [#445](https://github.com/fastly/terraform-provider-fastly/pull/445)
- Replace `StateFunc` with `ValidateDiagFunc` [#439](https://github.com/fastly/terraform-provider-fastly/pull/439)

BUG FIXES:

- Don't use `ParallelTest` for `no_auth` data source [#449](https://github.com/fastly/terraform-provider-fastly/pull/449)
- Introduce `no_auth` provider option [#444](https://github.com/fastly/terraform-provider-fastly/pull/444)
- Suppress gzip diff unless fields are explicitly set [#441](https://github.com/fastly/terraform-provider-fastly/pull/441)
- Fix parsing of log-levels by removing date/time prefix [#440](https://github.com/fastly/terraform-provider-fastly/pull/440)
- Fix bug with `fastly_tls_subscription` multi-SAN challenge [#435](https://github.com/fastly/terraform-provider-fastly/pull/435)
- Output variable refresh bug [#388](https://github.com/fastly/terraform-provider-fastly/pull/388)
- Use correct 'shield' value [#437](https://github.com/fastly/terraform-provider-fastly/pull/437)
- Fix `default_host` not being removed [#434](https://github.com/fastly/terraform-provider-fastly/pull/434)
- In `fastly_waf_rules` data source, request rule revisions from API [#428](https://github.com/fastly/terraform-provider-fastly/pull/428)

## 0.32.0 (June 17, 2021)

ENHANCEMENTS:

- Return 404 for non-existent service instead of a low-level nil entry error [#422](https://github.com/fastly/terraform-provider-fastly/pull/422)

BUG FIXES:

- Fix runtime panic in request-settings caused by incorrect type cast [#424](https://github.com/fastly/terraform-provider-fastly/pull/424)
- When `activate=true`, always read and clone from the active version [#423](https://github.com/fastly/terraform-provider-fastly/pull/423)

## 0.31.0 (June 14, 2021)

ENHANCEMENTS:

- Add support for ACL and extra redundancy options in S3 logging block [#417](https://github.com/fastly/terraform-provider-fastly/pull/417)
- Update default initial value for health check [#414](https://github.com/fastly/terraform-provider-fastly/pull/414)

BUG FIXES:

- Only set `cloned_version` after the version has been successfully validated [#418](https://github.com/fastly/terraform-provider-fastly/pull/418)

## 0.30.0 (May 12, 2021)

ENHANCEMENTS:

- Add director support for compute resource [#410](https://github.com/fastly/terraform-provider-fastly/pull/410)

## 0.29.1 (May 7, 2021)

BUG FIXES:

- Fix Header resource key names [#407](https://github.com/fastly/terraform-provider-fastly/pull/407)

## 0.29.0 (May 4, 2021)

ENHANCEMENTS:

- Add support for `file_max_bytes` configuration for Azure logging endpoint [#398](https://github.com/fastly/terraform-provider-fastly/pull/398)
- Support usage of IAM role in S3 and Kinesis logging endpoints [#403](https://github.com/fastly/terraform-provider-fastly/pull/403)
- Add support for `compression_codec` to logging file sink endpoints [#402](https://github.com/fastly/terraform-provider-fastly/pull/402)

DOCUMENTATION:

- Update debug mode instructions for Terraform 0.12.x [#405](https://github.com/fastly/terraform-provider-fastly/pull/405)

OTHER:

- Replace `master` with `main`. [#404](https://github.com/fastly/terraform-provider-fastly/pull/404)

## 0.28.2 (April 9, 2021)

BUG FIXES:

- Catch case where state from older version could be unexpected [#396](https://github.com/fastly/terraform-provider-fastly/pull/396)

## 0.28.1 (April 8, 2021)

BUG FIXES:

- Clone from `cloned_version` not `active_version` [#390](https://github.com/fastly/terraform-provider-fastly/pull/390)

## 0.28.0 (April 6, 2021)

ENHANCEMENTS:

- PATCH endpoint for TLS subscriptions [#370](https://github.com/fastly/terraform-provider-fastly/pull/370)
- Ensure passwords are marked as sensitive [#389](https://github.com/fastly/terraform-provider-fastly/pull/389)
- Add debug mode [#386](https://github.com/fastly/terraform-provider-fastly/pull/386)

BUG FIXES:

- Fix custom TLS configuration incorrectly omitting DNS records data [#392](https://github.com/fastly/terraform-provider-fastly/pull/392)
- Fix backend diff output incorrectly showing multiple resources being updated [#387](https://github.com/fastly/terraform-provider-fastly/pull/387)

DOCUMENTATION:

- Terraform 0.12+ no longer uses interpolation syntax for non-constant expressions [#384](https://github.com/fastly/terraform-provider-fastly/pull/384)

## 0.27.0 (March 16, 2021)

ENHANCEMENTS:

- Terraform Plugin SDK Upgrade [#379](https://github.com/fastly/terraform-provider-fastly/pull/379)
- Automate developer override for testing locally built provider [#382](https://github.com/fastly/terraform-provider-fastly/pull/382)

## 0.26.0 (March 5, 2021)

ENHANCEMENTS:

- Better sensitive value handling in Google pub/sub logging provider [#376](https://github.com/fastly/terraform-provider-fastly/pull/376)

BUG FIXES:

- Fix panic caused by incorrect type assert [#377](https://github.com/fastly/terraform-provider-fastly/pull/377)

## 0.25.0 (February 26, 2021)

ENHANCEMENTS:

- Add TLSCLientCert and TLSClientKey options for splunk logging ([#353](https://github.com/fastly/terraform-provider-fastly/pull/353))
- Add Dictionary to Compute service ([#361](https://github.com/fastly/terraform-provider-fastly/pull/361))
- Resources for Custom TLS and Platform TLS products ([#364](https://github.com/fastly/terraform-provider-fastly/pull/364))
- Managed TLS Subscriptions Resources ([#365](https://github.com/fastly/terraform-provider-fastly/pull/365))
- Ensure schema.Set uses custom SetDiff algorithm ([#366](https://github.com/fastly/terraform-provider-fastly/pull/366))
- Test speedup ([#371](https://github.com/fastly/terraform-provider-fastly/pull/371))
- Add service test sweeper ([#373](https://github.com/fastly/terraform-provider-fastly/pull/373))
- Add force_destroy flag to ACLs and Dicts to allow deleting non-empty lists ([#372](https://github.com/fastly/terraform-provider-fastly/pull/372))

## 0.24.0 (February 4, 2021)

ENHANCEMENTS:

- CI: check if docs need to be regenerated ([#362](https://github.com/fastly/terraform-provider-fastly/pull/362))
- Update go-fastly dependency to 3.0.0 ([#359](https://github.com/fastly/terraform-provider-fastly/pull/359))

BUG FIXES:

- Replace old doc generation process with new tfplugindocs tool ([#356](https://github.com/fastly/terraform-provider-fastly/pull/356))

## 0.23.0 (January 14, 2021)

ENHANCEMENT:

- Add support for kafka endpoints with sasl options ([#342](https://github.com/fastly/terraform-provider-fastly/pull/342))

## 0.22.0 (January 08, 2021)

ENHANCEMENT:

- Add Kinesis logging support ([#351](https://github.com/fastly/terraform-provider-fastly/pull/351))

## 0.21.3 (January 04, 2021)

NOTES:

- provider: Change version of go-fastly to v2.0.0 ([#341](https://github.com/fastly/terraform-provider-fastly/pull/341))

## 0.21.2 (December 16, 2020)

BUG FIXES:

- resource/fastly_service\_\*: Ensure we still refresh remote state when `activate` is set to `false` ([#345](https://github.com/fastly/terraform-provider-fastly/pull/345))

## 0.21.1 (October 15, 2020)

BUG FIXES:

- resource/fastly_service_waf_configuration: Guard `rule_exclusion` read to ensure API is only called if used. ([#330](https://github.com/fastly/terraform-provider-fastly/pull/330))

## 0.21.0 (October 14, 2020)

ENHANCEMENTS:

- resource/fastly_service_waf_configuration: Add `rule_exclusion` block which allows rules to be excluded from the WAF configuration. ([#328](https://github.com/fastly/terraform-provider-fastly/pull/328))

NOTES:

- provider: Change version of go-fastly to v2.0.0-alpha.1 ([#327](https://github.com/fastly/terraform-provider-fastly/pull/327))

## 0.20.4 (September 30, 2020)

NOTES:

- resource/fastly_service_acl_entries_v1: Change ACL documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
- resource/fastly_service_dictionary_items_v1: Change Dictionary documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
- resource/fastly_service_dynamic_snippet_content_v1: Change Dynamic Snippet documentation examples to use `for_each` attributes instead of `for` expressions. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/324))
- resource/fastly_service_waf_configuration: Correctly mark `allowed_request_content_type_charset` as optional in documentation. ([#324](https://github.com/fastly/terraform-provider-fastly/pull/322))

## 0.20.3 (September 23, 2020)

BUG FIXES:

- resource/fastly_service_v1/bigquerylogging: Ensure BigQuery logging `email`, `secret_key` fields are required and not optional. ([#319](https://github.com/fastly/terraform-provider-fastly/pull/319))

## 0.20.2 (September 22, 2020)

BUG FIXES:

- resource/fastly_service_v1: Improve performance of service read and delete logic. ([#311](https://github.com/fastly/terraform-provider-fastly/pull/311))
- resource/fastly_service_v1/logging_scalyr: Ensure `token` field is `sensitive` and thus hidden from plan. ([#310](https://github.com/fastly/terraform-provider-fastly/pull/310))

NOTES:

- resource/fastly_service_v1/s3logging: Document `server_side_encryption` and `server_side_encryption_kms_key_id` fields. ([#317](https://github.com/fastly/terraform-provider-fastly/pull/310))

## 0.20.1 (September 2, 2020)

BUG FIXES:

- resource/fastly_service_v1/backend: Ensure changes to backend fields result in updates instead of destroy and recreate. ([#304](https://github.com/fastly/terraform-provider-fastly/pull/304))
- resource/fastly_service_v1/logging\_\*: Fix logging acceptance tests by ensuring formatVersion in VCLLoggingAttributes is a \*uint. ([#307](https://github.com/fastly/terraform-provider-fastly/pull/307))

NOTES:

- provider: Add a [CONTRIBUTING.md](https://github.com/fastly/terraform-provider-fastly/blob/main/CONTRIBUTING.md) containing contributing guidelines and documentation. ([#305](https://github.com/fastly/terraform-provider-fastly/pull/307))

## 0.20.0 (August 10, 2020)

FEATURES:

- **New Data Source:** `fastly_waf_rules` Use this data source to fetch Fastly WAF rules and pass to a `fastly_service_waf_configuration`. ([#291](https://github.com/fastly/terraform-provider-fastly/pull/291))
- **New Resource:** `fastly_service_waf_configuration` Provides a Web Application Firewall configuration and rules that can be applied to a service. ([#291](https://github.com/fastly/terraform-provider-fastly/pull/291))

ENHANCEMENTS:

- resource/fastly_service_v1/waf: Add `waf` block to enable and configure a Web Application Firewall on a service. ([#285](https://github.com/fastly/terraform-provider-fastly/pull/285))

## 0.19.3 (July 30, 2020)

NOTES:

- provider: Initial release to the [Terraform Registry](https://registry.terraform.io/)

## 0.19.2 (July 22, 2020)

NOTES:

- resource/fastly_service_compute: Fixes resource references in website documentation ([#296](https://github.com/fastly/terraform-provider-fastly/pull/296))

## 0.19.1 (July 22, 2020)

NOTES:

- resource/fastly_service_compute: Update website documentation for compute resource to include correct terminology ([#294](https://github.com/fastly/terraform-provider-fastly/pull/294))

## 0.19.0 (July 22, 2020)

FEATURES:

- **New Resource:** `fastly_service_compute` ([#281](https://github.com/fastly/terraform-provider-fastly/pull/281))

ENHANCEMENTS:

- resource/fastly_service_compute: Add support for all logging providers ([#285](https://github.com/fastly/terraform-provider-fastly/pull/285))
- resource/fastly_service_compute: Add support for importing compute services ([#286](https://github.com/fastly/terraform-provider-fastly/pull/286))
- resource/fastly_service_v1/ftp_logging: Add support for `message_type` field to FTP logging endpoint ([#288](https://github.com/fastly/terraform-provider-fastly/pull/288))

BUG FIXES:

- resource/fastly_service_v1/s3logging: Fix error check which was causing a runtime panic with s3logging ([#290](https://github.com/fastly/terraform-provider-fastly/pull/290))

NOTES:

- provider: Update `go-fastly` client to v1.16.2 ([#288](https://github.com/fastly/terraform-provider-fastly/pull/288))
- provider: Refactor documentation templating and compilation ([#283](https://github.com/fastly/terraform-provider-fastly/pull/283))

## 0.18.0 (July 01, 2020)

ENHANCEMENTS:

- resource/fastly_service_v1/logging_digitalocean: Add DigitalOcean Spaces logging support ([#276](https://github.com/fastly/terraform-provider-fastly/pull/276))
- resource/fastly_service_v1/logging_cloudfiles: Add Rackspace Cloud Files logging support ([#275](https://github.com/fastly/terraform-provider-fastly/pull/275))
- resource/fastly_service_v1/logging_openstack: Add OpenStack logging support ([#273](https://github.com/fastly/terraform-provider-fastly/pull/274))
- resource/fastly_service_v1/logging_logshuttle: Add Log Shuttle logging support ([#273](https://github.com/fastly/terraform-provider-fastly/pull/273))
- resource/fastly_service_v1/logging_honeycomb: Add Honeycomb logging support ([#272](https://github.com/fastly/terraform-provider-fastly/pull/272))
- resource/fastly_service_v1/logging_heroku: Add Heroku logging support ([#271](https://github.com/fastly/terraform-provider-fastly/pull/271))

NOTES:

- resource/fastly_service_v1/\*: "GZIP" -> "Gzip" ([#279](https://github.com/fastly/terraform-provider-fastly/pull/279))
- resource/fastly_service_v1/logging_sftp: Update SFTP logging to use `ValidateFunc` for validating the `message_type` field ([#278](https://github.com/fastly/terraform-provider-fastly/pull/278))
- resource/fastly_service_v1/gcslogging: Update GCS logging to use `ValidateFunc` for validating the `message_type` field ([#278](https://github.com/fastly/terraform-provider-fastly/pull/278))
- resource/fastly_service_v1/blobstoragelogging: Update Azure Blob Storage logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))
- resource/fastly_service_v1/logging_ftp: Update FTP logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))
- resource/fastly_service_v1/s3logging: Update S3 logging to use a custom `StateFunc` for trimming whitespace from the `public_key` field ([#277](https://github.com/fastly/terraform-provider-fastly/pull/277))

## 0.17.1 (June 24, 2020)

NOTES:

- resource/fastly_service_v1/\*: Migrates service resources to implement the `ServiceAttributeDefinition` `interface` ([#269](https://github.com/fastly/terraform-provider-fastly/pull/269))

## 0.17.0 (June 22, 2020)

ENHANCEMENTS:

- resource/fastly_service_v1/logging_googlepubsub: Add Google Cloud Pub/Sub logging support ([#258](https://github.com/fastly/terraform-provider-fastly/pull/258))
- resource/fastly_service_v1/logging_kafka: Add Kafka logging support ([#254](https://github.com/fastly/terraform-provider-fastly/pull/254))
- resource/fastly_service_v1/logging_scalyr: Add Scalyr logging support ([#252](https://github.com/fastly/terraform-provider-fastly/pull/252))
- resource/fastly_service_v1/s3logging: Add support for public key field ([#249](https://github.com/fastly/terraform-provider-fastly/pull/249))
- resource/fastly_service_v1/logging_newrelic: Add New Relic logging support ([#243](https://github.com/fastly/terraform-provider-fastly/pull/243))
- resource/fastly_service_v1/logging_datadog: Add Datadog logging support ([#242](https://github.com/fastly/terraform-provider-fastly/pull/242))
- resource/fastly_service_v1/logging_loggly: Add Loggly logging support ([#241](https://github.com/fastly/terraform-provider-fastly/pull/241))
- resource/fastly_service_v1/logging_sftp: Add SFTP logging support ([#236](https://github.com/fastly/terraform-provider-fastly/pull/236))
- resource/fastly_service_v1/logging_ftp: Add FTP logging support ([#235](https://github.com/fastly/terraform-provider-fastly/pull/235))
- resource/fastly_service_v1/logging_elasticsearch: Add Elasticsearch logging support ([#234](https://github.com/fastly/terraform-provider-fastly/pull/234))

NOTES:

- resource/fastly_service_v1/sftp: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `public_key` and `secret_key` fields ([#268](https://github.com/fastly/terraform-provider-fastly/pull/268))
- resource/fastly_service_v1/bigquerylogging: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `secret_key` field ([#268](https://github.com/fastly/terraform-provider-fastly/pull/268))
- resource/fastly_service_v1/httpslogging: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `tls_ca_cert`, `tls_client_cert` and `tls_client_key` fields ([#264](https://github.com/fastly/terraform-provider-fastly/pull/264))
- resource/fastly_service_v1/splunk: Use `trimSpaceStateFunc` to trim leading and trailing whitespace from the `tls_ca_cert` field ([#264](https://github.com/fastly/terraform-provider-fastly/pull/264))
- resource/fastly_service_v1/\*: Migrate schemas to block separate block files ([#262](https://github.com/fastly/terraform-provider-fastly/pull/262))
- resource/fastly_service_v1/acl: Migrated to block file ([#253](https://github.com/fastly/terraform-provider-fastly/pull/253))
- provider: Update `go-fastly` client to v1.5.0 ([#248](https://github.com/fastly/terraform-provider-fastly/pull/248))

## 0.16.1 (June 03, 2020)

BUG FIXES:

- resource/fastly_service_v1/s3logging: Fix persistence of `server_side_encryption` and `server_side_encryption_kms_key_id` arguments ([#246](https://github.com/fastly/terraform-provider-fastly/pull/246))

## 0.16.0 (June 01, 2020)

ENHANCEMENTS:

- data-source/fastly_ip_ranges: Expose Fastly's IpV6 CIDR ranges via `ipv6_cidr_blocks` property ([#201](https://github.com/fastly/terraform-provider-fastly/pull/240))

NOTES:

- provider: Update `go-fastly` client to v1.14.1 ([#184](https://github.com/fastly/terraform-provider-fastly/pull/240))

## 0.15.0 (April 28, 2020)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `httpslogging` argument ([#222](https://github.com/fastly/terraform-provider-fastly/pull/222))
- resource/fastly_service_v1/splunk: Add `tls_hostname` and `tls_ca_cert` arguments ([#221](https://github.com/fastly/terraform-provider-fastly/pull/221))

NOTES:

- provider: Update `go-fastly` client to v1.10.0 ([#220](https://github.com/fastly/terraform-provider-fastly/pull/220))

## 0.14.0 (April 14, 2020)

FEATURES:

- **New Resource:** `fastly_user_v1` ([#214](https://github.com/fastly/terraform-provider-fastly/pull/214))

BUG FIXES:

- resource/fastly_service_v1/snippet: Fix support for `hash` snippet type ([#217](https://github.com/fastly/terraform-provider-fastly/pull/217))

NOTES:

- provider: Update `go` to v1.14.x ([#215](https://github.com/fastly/terraform-provider-fastly/pull/215))

## 0.13.0 (April 01, 2020)

ENHANCEMENTS:

- resource/fastly_service_v1/s3logging: Add `server_side_encryption` and `server_side_encryption_kms_key_id` arguments ([#206](https://github.com/fastly/terraform-provider-fastly/pull/206))
- resource/fastly_service_v1/snippet: Support `hash` in `type` validation ([#211](https://github.com/fastly/terraform-provider-fastly/issues/211))
- resource/fastly_service_v1/dynamicsnippet: Support `hash` in `type` validation ([#211](https://github.com/fastly/terraform-provider-fastly/issues/211))

NOTES:

- provider: Update `go-fastly` client to v1.7.2 ([#213](https://github.com/fastly/terraform-provider-fastly/pull/213))

## 0.12.1 (January 23, 2020)

BUG FIXES:

- resource/fastly_service_v1: Allow a service to be created with a `default_ttl` of `0` ([#205](https://github.com/fastly/terraform-provider-fastly/pull/205))

## 0.12.0 (January 21, 2020)

ENHANCEMENTS:

- resource/fastly_service_v1/syslog: Add `tls_client_cert` and `tls_client_key` arguments ([#203](https://github.com/fastly/terraform-provider-fastly/pull/203))

## 0.11.1 (December 16, 2019)

BUG FIXES:

- data-source/fastly_ip_ranges: Use `go-fastly` client in order to fetch Fastly's assigned IP ranges ([#201](https://github.com/fastly/terraform-provider-fastly/pull/201))

## 0.11.0 (October 15, 2019)

ENHANCEMENTS:

- resource/fastly_service_v1/dictionary: Add `write_only` argument ([#189](https://github.com/fastly/terraform-provider-fastly/pull/189))

NOTES:

- provider: The underlying Terraform codebase dependency for the provider SDK and acceptance testing framework has been migrated from `github.com/hashicorp/terraform` to `github.com/hashicorp/terraform-plugin-sdk`. They are functionality equivalent and this should only impact codebase development to switch imports. For more information see the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html). ([#191](https://github.com/fastly/terraform-provider-fastly/pull/191))
- provider: The actual Terraform version used by the provider will now be included in the `User-Agent` header for Terraform 0.12 and later. Terraform 0.11 and earlier will use `Terraform/0.11+compatible` as this information was not accessible in those versions. ([#182](https://github.com/fastly/terraform-provider-fastly/pull/182))

## 0.10.0 (October 02, 2019)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `cloned_version` argument ([#190](https://github.com/fastly/terraform-provider-fastly/pull/190))

## 0.9.0 (August 07, 2019)

FEATURES:

- **New Resource:** `fastly_service_acl_entries_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
- **New Resource:** `fastly_service_dictionary_items_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
- **New Resource:** `fastly_service_dynamic_snippet_content_v1` ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

ENHANCEMENTS:

- resource/fastly_service_v1: Add `acl` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
- resource/fastly_service_v1: Add `dictionary` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))
- resource/fastly_service_v1: Add `dynamicsnippet` argument ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

NOTES:

- provider: Update `go-fastly` client to v1.2.1 ([#184](https://github.com/fastly/terraform-provider-fastly/pull/184))

## 0.8.1 (July 12, 2019)

BUG FIXES:

- resource/fastly_service_v1/condition: Support `PREFETCH` in `type` validation ([#171](https://github.com/fastly/terraform-provider-fastly/issues/171))

## 0.8.0 (June 28, 2019)

NOTES:

- provider: This release includes only a Terraform SDK upgrade with compatibility for Terraform v0.12. The provider remains backwards compatible with Terraform v0.11 and there should not be any significant behavioural changes. ([#173](https://github.com/fastly/terraform-provider-fastly/pull/173))

## 0.7.0 (June 25, 2019)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `splunk` argument ([#130](https://github.com/fastly/terraform-provider-fastly/issues/130))
- resource/fastly_service_v1: Add `blobstoragelogging` argument ([#117](https://github.com/fastly/terraform-provider-fastly/issues/117))
- resource/fastly_service_v1: Add `comment` argument ([#70](https://github.com/fastly/terraform-provider-fastly/issues/70))
- resource/fastly_service_v1: Add `version_comment` argument ([#126](https://github.com/fastly/terraform-provider-fastly/issues/126))
- resource/fastly_service_v1/backend: Add `override_host` argument ([#163](https://github.com/fastly/terraform-provider-fastly/issues/163))
- resource/fastly_service_v1/condition: Add validation for `type` argument ([#148](https://github.com/fastly/terraform-provider-fastly/issues/148))

NOTES:

- provider: Update `go-fastly` client to v1.0.0 ([#165](https://github.com/fastly/terraform-provider-fastly/pull/165))

## 0.6.1 (May 29, 2019)

NOTES:

- provider: Switch codebase dependency management from `govendor` to Go modules ([#128](https://github.com/fastly/terraform-provider-fastly/pull/128))
- provider: Update `go-fastly` client to v0.4.3 ([#154](https://github.com/fastly/terraform-provider-fastly/pull/154))

## 0.6.0 (February 08, 2019)

ENHANCEMENTS:

- provider: Enable request/response logging ([#120](https://github.com/fastly/terraform-provider-fastly/issues/120))
- resource/fastly_service_v1: Add `activate` argument ([#45](https://github.com/fastly/terraform-provider-fastly/pull/45))

## 0.5.0 (January 08, 2019)

ENHANCEMENTS:

- resource/fastly_service_v1/s3logging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/papertrail: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/sumologic: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/gcslogging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/bigquerylogging: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/syslog: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))
- resource/fastly_service_v1/logentries: Add `placement` argument ([#106](https://github.com/fastly/terraform-provider-fastly/pull/106))

BUG FIXES:

- resource/fastly_service_v1/snippet: Exclude dynamic snippets ([#107](https://github.com/fastly/terraform-provider-fastly/pull/107))

## 0.4.0 (October 02, 2018)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `snippet` argument ([#93](https://github.com/fastly/terraform-provider-fastly/pull/93))
- resource/fastly_service_v1: Add `director` argument ([#43](https://github.com/fastly/terraform-provider-fastly/pull/43))
- resource/fastly_service_v1/bigquerylogging: Add `template` argument ([#90](https://github.com/fastly/terraform-provider-fastly/pull/90))

BUG FIXES:

- resource/fastly_service_v1: Handle deletion of already deleted or never created resources ([#89](https://github.com/fastly/terraform-provider-fastly/pull/89))

## 0.3.0 (August 02, 2018)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `bigquerylogging` argument ([#80](https://github.com/fastly/terraform-provider-fastly/issues/80))

## 0.2.0 (June 04, 2018)

ENHANCEMENTS:

- resource/fastly_service_v1/s3logging: Add `redundancy` argument ([64](https://github.com/fastly/terraform-provider-fastly/pull/64))
- provider: Support for overriding base API URL ([68](https://github.com/fastly/terraform-provider-fastly/pull/68))
- provider: Support for overriding user agent ([62](https://github.com/fastly/terraform-provider-fastly/pull/62))

BUG FIXES:

- resource/fastly_service_v1/sumologic: Properly detect changes and update resource ([56](https://github.com/fastly/terraform-provider-fastly/pull/56))

## 0.1.4 (January 16, 2018)

ENHANCEMENTS:

- resource/fastly_service_v1/s3logging: Add StateFunc to hash secrets ([#63](https://github.com/fastly/terraform-provider-fastly/issues/63))

## 0.1.3 (December 18, 2017)

ENHANCEMENTS:

- resource/fastly_service_v1: Add `logentries` argument ([#24](https://github.com/fastly/terraform-provider-fastly/issues/24))
- resource/fastly_service_v1: Add `syslog` argument ([#16](https://github.com/fastly/terraform-provider-fastly/issues/16))

ENHANCEMENTS:

- resource/fastly_service_v1/syslog: Add `message_type` argument ([#30](https://github.com/fastly/terraform-provider-fastly/issues/30))

## 0.1.2 (August 02, 2017)

ENHANCEMENTS:

- resource/fastly_service_v1/backend: Add `ssl_ca_cert` argument ([#11](https://github.com/fastly/terraform-provider-fastly/issues/11))
- resource/fastly_service_v1/s3logging: Add `message_type` argument ([#14](https://github.com/fastly/terraform-provider-fastly/issues/14))
- resource/fastly_service_v1/gcslogging: Add environment variable support for `secret_key` argument ([#15](https://github.com/fastly/terraform-provider-fastly/issues/15))

BUG FIXES:

- resource/fastly_service_v1/s3logging: Update default value of `domain` argument ([#12](https://github.com/fastly/terraform-provider-fastly/issues/12))

## 0.1.1 (June 21, 2017)

NOTES:

- provider: Bumping the provider version to get around provider caching issues - still same functionality

## 0.1.0 (June 20, 2017)

NOTES:

- provider: Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
