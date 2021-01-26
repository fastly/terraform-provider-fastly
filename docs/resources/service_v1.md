---
page_title: "Fastly: service_v1"
---

# fastly_service_v1

Provides a Fastly Service, representing the configuration for a website, app,
API, or anything else to be served through Fastly. A Service encompasses Domains
and Backends.

The Service resource requires a domain name that is correctly set up to direct
traffic to the Fastly service. See Fastly's guide on [Adding CNAME Records][fastly-cname]
on their documentation site for guidance.

## Example Usage

Basic usage:

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}
```

Basic usage with an Amazon S3 Website and that removes the `x-amz-request-id` header:

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
    override_host = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
  }

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  gzip {
    name          = "file extensions and content types"
    extensions    = ["css", "js"]
    content_types = ["text/html", "text/css"]
  }

  force_destroy = true
}

resource "aws_s3_bucket" "website" {
  bucket = "demo.notexample.com"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
```

Basic usage with [custom
VCL](https://docs.fastly.com/vcl/custom-vcl/uploading-custom-vcl/):

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true

  vcl {
    name    = "my_custom_main_vcl"
    content = "${file("${path.module}/my_custom_main.vcl")}"
    main    = true
  }

  vcl {
    name    = "my_custom_library_vcl"
    content = "${file("${path.module}/my_custom_library.vcl")}"
  }
}
```

Basic usage with [custom Director](https://developer.fastly.com/reference/api/load-balancing/directors/director/):

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  backend {
    address = "127.0.0.2"
    name    = "origin2"
    port    = 80
  }

  director {
    name = "mydirector"
    quorum = 0
    type = 3
    backends = [ "origin1", "origin2" ]
  }

  force_destroy = true
}
```

-> **Note:** The following example is only available from 0.20.0 of the Fastly Terraform provider.

Basic usage with [Web Application Firewall](https://developer.fastly.com/reference/api/waf/):

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "example.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "WAF_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # This condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic.
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name              = "WAF_Response"
    status            = "403"
    response          = "Forbidden"
    content_type      = "text/html"
    content           = "<html><body>Forbidden</body></html>"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "WAF_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}
```

-> **Note:** For an AWS S3 Bucket, the Backend address is
`<domain>.s3-website-<region>.amazonaws.com`. The `override_host` attribute
should be set to `<bucket_name>.s3-website-<region>.amazonaws.com` in the `backend` block. See the
Fastly documentation on [Amazon S3][fastly-s3].

## Argument Reference

The following arguments are supported:

* `activate` - (Optional) Conditionally prevents the Service from being activated. The apply step will continue to create a new draft version but will not activate it if this is set to false. Default true.
* `name` - (Required) The unique name for the Service to create.
* `comment` - (Optional) Description field for the service. Default `Managed by Terraform`.
* `version_comment` - (Optional) Description field for the version.
* `domain` - (Required) A set of Domain names to serve as entry points for your
Service. [Defined below](#domain-block).
* `backend` - (Optional) A set of Backends to service requests from your Domains.
[Defined below](#backend-block). Backends must be defined in this argument, or defined in the
`vcl` argument below
* `condition` - (Optional) A set of conditions to add logic to any basic
configuration object in this service. [Defined below](#condition-block).
* `cache_setting` - (Optional) A set of Cache Settings, allowing you to override.
[Defined below](#cache_setting-block).
* `director` - (Optional) A director to allow more control over balancing traffic over backends.
when an item is not to be cached based on an above `condition`.
[Defined below](#director-block)
* `gzip` - (Required) A set of gzip rules to control automatic gzipping of
content. [Defined below](#gzip-block).
* `header` - (Optional) A set of Headers to manipulate for each request.
[Defined below](#header-block).
* `healthcheck` - (Optional) Automated healthchecks on the cache that can change how Fastly interacts with the cache based on its health.
[Defined below](#healthcheck-block).
* `default_host` - (Optional) The default hostname.
* `default_ttl` - (Optional) The default Time-to-live (TTL) for
requests.
* `force_destroy` - (Optional) Services that are active cannot be destroyed. In
order to destroy the Service, set `force_destroy` to `true`. Default `false`.
* `request_setting` - (Optional) A set of Request modifiers.
[Defined below](#request_setting-block)
* `s3logging` - (Optional) A set of S3 Buckets to send streaming logs too.
[Defined below](#s3logging-block).
* `papertrail` - (Optional) A Papertrail endpoint to send streaming logs too.
[Defined below](#papertrail-block).
* `sumologic` - (Optional) A Sumologic endpoint to send streaming logs too.
[Defined below](#sumologic-block).
* `gcslogging` - (Optional) A gcs endpoint to send streaming logs too.
[Defined below](#gcslogging-block).
* `bigquerylogging` - (Optional) A BigQuery endpoint to send streaming logs too.
[Defined below](#bigquerylogging-block).
* `syslog` - (Optional) A syslog endpoint to send streaming logs too.
[Defined below](#syslog-block).
* `logentries` - (Optional) A logentries endpoint to send streaming logs too.
[Defined below](#logentries-block).
* `splunk` - (Optional) A Splunk endpoint to send streaming logs too.
[Defined below](#splunk-block).
* `blobstoragelogging` - (Optional) An Azure Blob Storage endpoint to send streaming logs too.
[Defined below](#blobstoragelogging-block).
* `httpslogging` - (Optional) An HTTPS endpoint to send streaming logs to.
[Defined below](#httpslogging-block).
* `logging_elasticsearch` - (optional) An Elasticsearch endpoint to send streaming logs to.
[Defined below](#logging_elasticsearch-block).
* `logging_ftp` - (Optional) An FTP endpoint to send streaming logs to.
[Defined below](#logging_ftp-block).
* `logging_sftp` - (Optional) An SFTP endpoint to send streaming logs to.
[Defined below](#logging_sftp-block).
* `logging_datadog` - (Optional) A Datadog endpoint to send streaming logs to.
[Defined below](#logging_datadog-block).
* `logging_loggly` - (Optional) A Loggly endpoint to send streaming logs to.
[Defined below](#logging_loggly-block).
* `logging_newrelic` - (Optional) A New Relic endpoint to send streaming logs to.
[Defined below](#logging_newrelic-block).
* `logging_scalyr` - (Optional) A Scalyr endpoint to send streaming logs to.
[Defined below](#logging_scalyr-block).
* `logging_googlepubsub` - (Optional) A Google Cloud Pub/Sub endpoint to send streaming logs to.
[Defined below](#logging_googlepubsub-block).
* `logging_kafka` - (Optional) A Kafka endpoint to send streaming logs to.
[Defined below](#logging_kafka-block).
* `logging_heroku` - (Optional) A Heroku endpoint to send streaming logs to.
[Defined below](#logging_heroku-block).
* `logging_honeycomb` - (Optional) A Honeycomb endpoint to send streaming logs to.
[Defined below](#logging_honeycomb-block).
* `logging_logshuttle` - (Optional) A Log Shuttle endpoint to send streaming logs to.
[Defined below](#logging_logshuttle-block).
* `logging_openstack` - (Optional) An OpenStack endpoint to send streaming logs to.
[Defined below](#logging_openstack-block).
* `logging_digitalocean` - (Optional) A DigitalOcean Spaces endpoint to send streaming logs to.
[Defined below](#logging_digitalocean-block).
* `logging_cloudfiles` - (Optional) A Rackspace Cloud Files endpoint to send streaming logs to.
[Defined below](#logging_cloudfiles-block).
* `logging_kinesis` - (Optional) A Kinesis endpoint to send streaming logs to.
[Defined below](#logging_kinesis-block).
* `response_object` - (Optional) Allows you to create synthetic responses that exist entirely on the varnish machine. Useful for creating error or maintenance pages that exists outside the scope of your datacenter. Best when used with Condition objects.
[Defined below](#response_object-block)
* `snippet` - (Optional) A set of custom, "regular" (non-dynamic) VCL Snippet configuration blocks.
[Defined below](#snippet-block).
* `dynamicsnippet` - (Optional) A set of custom, "dynamic" VCL Snippet configuration blocks.
[Defined below](#dynamicsnippet-block).
* `vcl` - (Optional) A set of custom VCL configuration blocks.
[Defined below](#vcl-block). See the [Fastly documentation](https://docs.fastly.com/vcl/custom-vcl/uploading-custom-vcl/) for more information on using custom VCL.
* `acl` - (Optional) A set of ACL configuration blocks.
[Defined below](#acl-block).
* `dictionary` - (Optional) A set of dictionaries that allow the storing of key values pair for use within VCL functions.
[Defined below](#dictionary-block).
* `waf` - (Optional) A WAF configuration block.
[Defined below](#waf-block).

### domain block

The `domain` block supports:

* `name` - (Required) The domain to which this Service will respond.
* `comment` - (Optional) An optional comment about the Domain.

### backend block

The `backend` block supports:

* `name` - (Required, string) Name for this Backend. Must be unique to this Service.
* `address` - (Required, string) An IPv4, hostname, or IPv6 address for the Backend.
* `auto_loadbalance` - (Optional, boolean) Denotes if this Backend should be
included in the pool of backends that requests are load balanced against.
Default `true`.
* `between_bytes_timeout` - (Optional) How long to wait between bytes in milliseconds. Default `10000`.
* `connect_timeout` - (Optional) How long to wait for a timeout in milliseconds.
Default `1000`
* `error_threshold` - (Optional) Number of errors to allow before the Backend is marked as down. Default `0`.
* `first_byte_timeout` - (Optional) How long to wait for the first bytes in milliseconds. Default `15000`.
* `max_conn` - (Optional) Maximum number of connections for this Backend.
Default `200`.
* `port` - (Optional) The port number on which the Backend responds. Default `80`.
* `override_host` - (Optional) The hostname to override the Host header.
* `request_condition` - (Optional, string) Name of already defined `condition`, which if met, will select this backend during a request.
* `use_ssl` - (Optional) Whether or not to use SSL to reach the backend. Default `false`.
* `max_tls_version` - (Optional) Maximum allowed TLS version on SSL connections to this backend.
* `min_tls_version` - (Optional) Minimum allowed TLS version on SSL connections to this backend.
* `ssl_ciphers` - (Optional) Comma separated list of OpenSSL Ciphers to try when negotiating to the backend.
* `ssl_ca_cert` - (Optional) CA certificate attached to origin.
* `ssl_client_cert` - (Optional) Client certificate attached to origin. Used when connecting to the backend.
* `ssl_client_key` - (Optional) Client key attached to origin. Used when connecting to the backend.
* `ssl_check_cert` - (Optional) Be strict about checking SSL certs. Default `true`.
* `ssl_hostname` - (Optional, deprecated by Fastly) Used for both SNI during the TLS handshake and to validate the cert.
* `ssl_cert_hostname` - (Optional) Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all.
* `ssl_sni_hostname` - (Optional) Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all.
* `shield` - (Optional) The POP of the shield designated to reduce inbound load. Valid values for `shield` are included in the [`GET /datacenters`](https://developer.fastly.com/reference/api/utils/datacenter/) API response.
* `weight` - (Optional) The [portion of traffic](https://docs.fastly.com/en/guides/load-balancing-configuration#how-weight-affects-load-balancing) to send to this Backend. Each Backend receives `weight / total` of the traffic. Default `100`.
* `healthcheck` - (Optional) Name of a defined `healthcheck` to assign to this backend.


### condition block

The `condition` block supports allows you to add logic to any basic configuration
object in a service. See Fastly's documentation
["About Conditions"](https://docs.fastly.com/en/guides/about-conditions)
for more detailed information on using Conditions. The Condition `name` can be
used in the `request_condition`, `response_condition`, or
`cache_condition` attributes of other block settings.

* `name` - (Required) The unique name for the condition.
* `statement` - (Required) The statement used to determine if the condition is met.
* `type` - (Required) Type of condition, either `REQUEST` (req), `RESPONSE`
(req, resp), or `CACHE` (req, beresp).
* `priority` - (Optional) A number used to determine the order in which multiple
conditions execute. Lower numbers execute first. Default `10`.

### director block

The `director` block supports:

* `name` - (Required) Unique name for this Director.
* `backends` - (Required) Names of defined backends to map the director to. Example: `[ "origin1", "origin2" ]`
* `comment` - (Optional) An optional comment about the Director.
* `shield` - (Optional) Selected POP to serve as a "shield" for backends. Valid values for `shield` are included in the [`GET /datacenters`](https://developer.fastly.com/reference/api/utils/datacenter/) API response.
* `capacity` - (Optional) Load balancing weight for the backends. Default `100`.
* `quorum` - (Optional) Percentage of capacity that needs to be up for the director itself to be considered up. Default `75`.
* `type` - (Optional) Type of load balance group to use. Integer, 1 to 4. Values: `1` (random), `3` (hash), `4` (client).  Default `1`.
* `retries` - (Optional) How many backends to search if it fails. Default `5`.

### cache_setting block

The `cache_setting` block supports:

* `name` - (Required) Unique name for this Cache Setting.
* `action` - (Optional) One of `cache`, `pass`, or `restart`, as defined
on Fastly's documentation under ["Caching action descriptions"](https://docs.fastly.com/en/guides/controlling-caching#caching-action-descriptions).
* `cache_condition` - (Optional) Name of already defined `condition` used to test whether this settings object should be used. This `condition` must be of type `CACHE`.
* `stale_ttl` - (Optional) Max "Time To Live" for stale (unreachable) objects.
* `ttl` - (Optional) The Time-To-Live (TTL) for the object.

### gzip block

The `gzip` block supports:

* `name` - (Required) A unique name.
* `content_types` - (Optional) The content-type for each type of content you wish to
have dynamically gzip'ed. Example: `["text/html", "text/css"]`.
* `extensions` - (Optional) File extensions for each file type to dynamically
gzip. Example: `["css", "js"]`.
* `cache_condition` - (Optional) Name of already defined `condition` controlling when this gzip configuration applies. This `condition` must be of type `CACHE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].


### header block

The `header` block supports adding, removing, or modifying Request and Response
headers. See Fastly's documentation on
[Adding or modifying headers on HTTP requests and responses](https://docs.fastly.com/en/guides/adding-or-modifying-headers-on-http-requests-and-responses#field-description-table) for more detailed information on any of the properties below.

* `name` - (Required) Unique name for this header attribute.
* `action` - (Required) The Header manipulation action to take; must be one of
`set`, `append`, `delete`, `regex`, or `regex_repeat`.
* `type` - (Required) The Request type on which to apply the selected Action; must be one of `request`, `fetch`, `cache` or `response`.
* `destination` - (Required) The name of the header that is going to be affected by the Action.
* `ignore_if_set` - (Optional) Do not add the header if it is already present. (Only applies to the `set` action.). Default `false`.
* `source` - (Optional) Variable to be used as a source for the header
content. (Does not apply to the `delete` action.)
* `regex` - (Optional) Regular expression to use (Only applies to the `regex` and `regex_repeat` actions.)
* `substitution` - (Optional) Value to substitute in place of regular expression. (Only applies to the `regex` and `regex_repeat` actions.)
* `priority` - (Optional) Lower priorities execute first. Default: `100`.
* `request_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `REQUEST`.
* `cache_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `CACHE`.
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].

### healthcheck block

The `healthcheck` block supports:

* `name` - (Required) A unique name to identify this Healthcheck.
* `host` - (Required) The Host header to send for this Healthcheck.
* `path` - (Required) The path to check.
* `check_interval` - (Optional) How often to run the Healthcheck in milliseconds. Default `5000`.
* `expected_response` - (Optional) The status code expected from the host. Default `200`.
* `http_version` - (Optional) Whether to use version 1.0 or 1.1 HTTP. Default `1.1`.
* `initial` - (Optional) When loading a config, the initial number of probes to be seen as OK. Default `2`.
* `method` - (Optional) Which HTTP method to use. Default `HEAD`.
* `threshold` - (Optional) How many Healthchecks must succeed to be considered healthy. Default `3`.
* `timeout` - (Optional) Timeout in milliseconds. Default `500`.
* `window` - (Optional) The number of most recent Healthcheck queries to keep for this Healthcheck. Default `5`.

### request_setting block

The `request_setting` block allow you to customize Fastly's request handling, by
defining behavior that should change based on a predefined `condition`:

* `name` - (Required) Unique name to refer to this Request Setting.
* `request_condition` - (Optional) Name of already defined `condition` to
determine if this request setting should be applied.
* `max_stale_age` - (Optional) How old an object is allowed to be to serve
`stale-if-error` or `stale-while-revalidate`, in seconds.
* `force_miss` - (Optional) Force a cache miss for the request. If specified,
can be `true` or `false`.
* `force_ssl` - (Optional) Forces the request to use SSL (Redirects a non-SSL request to SSL).
* `action` - (Optional) Allows you to terminate request handling and immediately
perform an action. When set it can be `lookup` or `pass` (Ignore the cache completely).
* `bypass_busy_wait` - (Optional) Disable collapsed forwarding, so you don't wait
for other objects to origin.
* `hash_keys` - (Optional) Comma separated list of varnish request object fields
that should be in the hash key.
* `xff` - (Optional) X-Forwarded-For, should be `clear`, `leave`, `append`,
`append_all`, or `overwrite`. Default `append`.
* `timer_support` - (Optional) Injects the X-Timer info into the request for
viewing origin fetch durations.
* `geo_headers` - (Optional) Injects Fastly-Geo-Country, Fastly-Geo-City, and
Fastly-Geo-Region into the request headers.
* `default_host` - (Optional) Sets the host header.

### s3logging block

The `s3logging` block supports:

* `name` - (Required) The unique name of the S3 logging endpoint.
* `bucket_name` - (Required) The name of the bucket in which to store the logs.
* `s3_access_key` - (Required) AWS Access Key of an account with the required
permissions to post logs. It is **strongly** recommended you create a separate
IAM user with permissions to only operate on this Bucket. This key will be
not be encrypted. You can provide this key via an environment variable, `FASTLY_S3_ACCESS_KEY`.
* `s3_secret_key` - (Required) AWS Secret Key of an account with the required
permissions to post logs. It is **strongly** recommended you create a separate
IAM user with permissions to only operate on this Bucket. This secret will be
not be encrypted. You can provide this secret via an environment variable, `FASTLY_S3_SECRET_KEY`.
* `path` - (Optional) Path to store the files. Must end with a trailing slash.
If this field is left empty, the files will be saved in the bucket's root path.
* `domain` - (Optional) If you created the S3 bucket outside of `us-east-1`,
then specify the corresponding bucket endpoint. Example: `s3-us-west-2.amazonaws.com`.
* `public_key` - (Optional) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `period` - (Optional) How frequently the logs should be transferred, in
seconds. Default `3600`.
* `gzip_level` - (Optional) Level of Gzip compression, from `0-9`. `0` is no
compression. `1` is fastest and least compressed, `9` is slowest and most
compressed. Default `0`.
* `message_type` - (Optional) How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`.  Default `classic`.
* `timestamp_format` - (Optional) `strftime` specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `redundancy` - (Optional) The S3 redundancy level. Should be formatted; one of: `standard`, `reduced_redundancy` or null. Default `null`.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (`%h %l %u %t %r %>s`).
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either 1 (the default, version 1 log format) or 2 (the version 2 log format).
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.
* `server_side_encryption` - (Optional) Specify what type of server side encryption should be used. Can be either `AES256` or `aws:kms`.
* `server_side_encryption_kms_key_id` - (Optional) Server-side KMS Key ID. Must be set if `server_side_encryption` is set to `aws:kms`.

### papertrail block

The `papertrail` block supports:

* `name` - (Required) A unique name to identify this Papertrail endpoint.
* `address` - (Required) The address of the Papertrail endpoint.
* `port` - (Required) The port associated with the address where the Papertrail endpoint can be accessed.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (`%h %l %u %t %r %>s`)
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### sumologic block

The `sumologic` block supports:

* `name` - (Required) A unique name to identify this Sumologic endpoint.
* `url` - (Required) The URL to Sumologic collector endpoint
* `message_type` - (Optional) How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`. Default `classic`. See [Fastly's Documentation on Sumologic][fastly-sumologic]
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (`%h %l %u %t %r %>s`)
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either 1 (the default, version 1 log format) or 2 (the version 2 log format).
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### gcslogging block

The `gcslogging` block supports:

* `name` - (Required) A unique name to identify this GCS endpoint.
* `email` - (Required) The email address associated with the target GCS bucket on your account. You may optionally provide this secret via an environment variable, `FASTLY_GCS_EMAIL`.
* `bucket_name` - (Required) The name of the bucket in which to store the logs.
* `secret_key` - (Required) The secret key associated with the target gcs bucket on your account. You may optionally provide this secret via an environment variable, `FASTLY_GCS_SECRET_KEY`. A typical format for the key is PEM format, containing actual newline characters where required.
* `path` - (Optional) Path to store the files. Must end with a trailing slash.
If this field is left empty, the files will be saved in the bucket's root path.
* `period` - (Optional) How frequently the logs should be transferred, in
seconds. Default `3600`.
* `gzip_level` - (Optional) Level of Gzip compression, from `0-9`. `0` is no
compression. `1` is fastest and least compressed, `9` is slowest and most
compressed. Default `0`.
* `message_type` - (Optional) How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`. Default `classic`. [Fastly Documentation](https://developer.fastly.com/reference/api/logging/gcs/)
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (`%h %l %u %t %r %>s`)
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### bigquerylogging block

The `bigquerylogging` block supports:

* `name` - (Required) A unique name to identify this BigQuery logging endpoint.
* `project_id` - (Required) The ID of your GCP project.
* `dataset` - (Required) The ID of your BigQuery dataset.
* `table` - (Required) The ID of your BigQuery table.
* `email` - (Required) The email for the service account with write access to your BigQuery dataset. If not provided, this will be pulled from a `FASTLY_BQ_EMAIL` environment variable.
* `secret_key` - (Required) The secret key associated with the sservice account that has write access to your BigQuery table. If not provided, this will be pulled from the `FASTLY_BQ_SECRET_KEY` environment variable. Typical format for this is a private key in a string with newlines.
* `format` - (Optional) Apache style log formatting. Must produce JSON that matches the schema of your BigQuery table.
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `template` - (Optional) Big query table name suffix template. If set will be interpreted as a strftime compatible string and used as the [Template Suffix for your table](https://cloud.google.com/bigquery/streaming-data-into-bigquery#template-tables).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### syslog block

The `syslog` block supports:

* `name` - (Required) A unique name to identify this Syslog endpoint.
* `address` - (Required) A hostname or IPv4 address of the Syslog endpoint.
* `port` - (Optional) The port associated with the address where the Syslog endpoint can be accessed. Default `514`.
* `token` - (Optional) Whether to prepend each message with a specific token.
* `use_tls` - (Optional) Whether to use TLS for secure logging. Default `false`.
* `tls_hostname` - (Optional) Used during the TLS handshake to validate the certificate.
* `tls_ca_cert` - (Optional) A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SYSLOG_CA_CERT`
* `tls_client_cert` - (Optional) The client certificate used to make authenticated requests. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SYSLOG_CLIENT_CERT`
* `tls_client_key` - (Optional) The client private key used to make authenticated requests. Must be in PEM format. You can provide this key via an environment variable, `FASTLY_SYSLOG_CLIENT_KEY`
* `message_type` - (Optional) How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`.  Default `classic`.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (%h %l %u %t %r %>s)
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either 1 (the default, version 1 log format) or 2 (the version 2 log format).
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### logentries block

The `logentries` block supports:

* `name` - (Required) A unique name to identify this GCS endpoint.
* `token` - (Required) Logentries Token to be used for authentication (https://logentries.com/doc/input-token/).
* `port` - (Optional) The port number configured in Logentries to send logs to. Defaults to `20000`.
* `use_tls` - (Optional) Whether to use TLS for secure logging. Defaults to `true`
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Defaults to Apache Common Log format (`%h %l %u %t %r %>s`).
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either 1 (the default, version 1 log format) or 2 (the version 2 log format).
* `response_condition` - (Optional) Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals][fastly-conditionals].
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### splunk block

The `splunk` block supports:

* `name` - (Required) A unique name to identify the Splunk endpoint.
* `url` - (Required) The Splunk URL to stream logs to.
* `token` - (Required) The Splunk token to be used for authentication.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Default `%h %l %u %t \"%r\" %>s %b`.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) Where in the generated VCL the logging call should be placed, overriding any `format_version` default. Can be either `none` or `waf_debug`.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.
* `tls_hostname` - (Optional) The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN).
* `tls_ca_cert` - (Optional) A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SPLUNK_CA_CERT`.

### blobstoragelogging block

The `blobstoragelogging` block supports:

* `name` - (Required) A unique name to identify the Azure Blob Storage endpoint.
* `account_name` - (Required) The unique Azure Blob Storage namespace in which your data objects are stored.
* `container` - (Required) The name of the Azure Blob Storage container in which to store logs.
* `sas_token` - (Required) The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work.
* `path` - (Optional) The path to upload logs to. Must end with a trailing slash. If this field is left empty, the files will be saved in the container's root path.
* `period` - (Optional) How frequently the logs should be transferred in seconds. Default `3600`.
* `timestamp_format` - (Optional) `strftime` specified timestamp formatting. Default `%Y-%m-%dT%H:%M:%S.000`.
* `gzip_level` - (Optional) Level of Gzip compression from `0`to `9`. `0` means no compression. `1` is the fastest and the least compressed version, `9` is the slowest and the most compressed version. Default `0`.
* `public_key` - (Optional) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `message_type` - (Optional) How the message should be formatted. Can be either `classic`, `loggly`, `logplex` or `blank`.  Default `classic`.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting. Default `%h %l %u %t \"%r\" %>s %b`.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) Where in the generated VCL the logging call should be placed, overriding any `format_version` default. Can be either `none` or `waf_debug`.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### httpslogging block

The `httpslogging` block supports:

* `name` - (Required) The unique name of the HTTPS logging endpoint.
* `url` - (Required) URL that log data will be sent to. Must use the https protocol.
* `request_max_entries` - (Optional) The maximum number of logs sent in one request.
* `request_max_bytes` - (Optional) The maximum number of bytes sent in one request.
* `content_type` - (Optional) Value of the `Content-Type` header sent with the request.
* `header_name` - (Optional) Custom header sent with the request.
* `header_value` - (Optional) Value of the custom header sent with the request.
* `method` - (Optional) HTTP method used for request. Can be either `POST` or `PUT`. Default `POST`.
* `json_format` - Formats log entries as JSON. Can be either disabled (`0`), array of json (`1`), or newline delimited json (`2`).
* `tls_hostname` - (Optional) Used during the TLS handshake to validate the certificate.
* `tls_ca_cert` - (Optional) A secure certificate to authenticate the server with. Must be in PEM format.
* `tls_client_cert` - (Optional) The client certificate used to make authenticated requests. Must be in PEM format.
* `tls_client_key` - (Optional) The client private key used to make authenticated requests. Must be in PEM format.
* `message_type` - How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`.  Default `blank`.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### logging_elasticsearch block

The `logging_elasticsearch` block supports:

* `name` - (Required) The unique name of the Elasticsearch logging endpoint.
* `url` - (Required) The Elasticsearch URL to stream logs to.
* `index` - (Required) The name of the Elasticsearch index to send documents (logs) to.
* `user` - (Optional) BasicAuth username for Elasticsearch.
* `password` - (Optional) BasicAuth password for Elasticsearch.
* `pipeline` - (Optional) The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing.
* `request_max_bytes` - (Optional) The maximum number of bytes sent in one request. Defaults to `0` for unbounded.
* `request_max_entries` - (Optional) The maximum number of logs sent in one request. Defaults to `0` for unbounded.
* `tls_ca_cert` - (Optional) A secure certificate to authenticate the server with. Must be in PEM format.
* `tls_client_cert` - (Optional) The client certificate used to make authenticated requests. Must be in PEM format.
* `tls_client_key` - (Optional) The client private key used to make authenticated requests. Must be in PEM format.
* `tls_hostname` - (Optional) The hostname used to verify the server's certificate. It can either be the Common Name (CN) or a Subject Alternative Name (SAN).
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### logging_ftp block

The `logging_ftp` block supports:

* `name` - (Required) The unique name of the FTP logging endpoint.
* `address` - (Required) The FTP address to stream logs to.
* `user` - (Required) The username for the server (can be `anonymous`).
* `password` - (Required) The password for the server (for anonymous use an email address).
* `path` - (Required) The path to upload log files to. If the path ends in `/` then it is treated as a directory.
* `port` - (Optional) The port number. Default: `21`.
* `gzip_level` - (Optional) Gzip Compression level. Default `0`.
* `period` - (Optional) How frequently the logs should be transferred, in seconds (Default 3600).
* `public_key` - (Optional) The PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `timestamp_format` - (Optional) specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.
* `response_condition` - (Optional) The name of the condition to apply.

### logging_sftp block

The `logging_sftp` block supports:

* `name` - (Required) The unique name of the SFTP logging endpoint.
* `address` - (Required) The SFTP address to stream logs to.
* `path` - (Required) The path to upload log files to. If the path ends in / then it is treated as a directory.
* `ssh_known_hosts` - (Required) A list of host keys for all hosts we can connect to over SFTP.
* `user` - (Required) The username for the server.
* `port` - (Optional) The port the SFTP service listens on. (Default: `22`).
* `password` - (Optional) The password for the server. If both `password` and `secret_key` are passed, `secret_key` will be preferred.
* `secret_key` - (Optional) The SSH private key for the server. If both `password` and `secret_key` are passed, `secret_key` will be preferred.
* `gzip_level` - (Optional) What level of Gzip encoding to have when dumping logs (default 0, no compression).
* `period` - (Optional) How frequently log files are finalized so they can be available for reading (in seconds, default `3600`).
* `public_key` - (Optional) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `timestamp_format` - (Optional) The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `message_type` - (Optional) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `response_condition` - (Optional) The name of the condition to apply.
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.

### logging_datadog block

The `logging_datadog` block supports:

* `name` - (Required) The unique name of the Datadog logging endpoint.
* `token` - (Required) The API key from your Datadog account.
* `region` - (Optional) The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.
* `response_condition` - (Optional) The name of the condition to apply.

### logging_loggly block

The `logging_loggly` block supports:

* `name` - (Required) The unique name of the Loggly logging endpoint.
* `token` - (Required) The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/).
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_newrelic block

The `logging_newrelic` block supports:

* `name` - (Required) The unique name of the New Relic logging endpoint.
* `token` - (Required) The Insert API key from the Account page of your New Relic account.
* `format` - (Optional) Apache style log formatting. Your log must produce valid JSON that New Relic Logs can ingest.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed.
* `response_condition` - (Optional) The name of the condition to apply.

### logging_scalyr block

The `logging_scalyr` block supports:

* `name` - (Required) The unique name of the Scalyr logging endpoint.
* `token` - (Required) The token to use for authentication (https://www.scalyr.com/keys).
* `region` - (Optional) The region that log data will be sent to. One of US or EU. Defaults to US if undefined.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### logging_googlepubsub block

The `logging_googlepubsub` block supports:

* `name` - (Required) The unique name of the Google Cloud Pub/Sub logging endpoint.
* `user` - (Required) Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON.
* `secret_key` - (Required) Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON.
* `project_id` - (Required) The ID of your Google Cloud Platform project.
* `topic` - (Required) The Google Cloud Pub/Sub topic to which logs will be published.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### logging_kafka block

The `logging_kafka` block supports:

* `name` - (Required) The unique name of the Kafka logging endpoint.
* `topic` - (Required) The Kafka topic to send logs to.
* `brokers` - (Required) A comma-separated list of IP addresses or hostnames of Kafka brokers.
* `compression_codec` - (Optional) The codec used for compression of your logs. One of: gzip, snappy, lz4.
* `required_acks` - (Optional) The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0 No servers need to respond. -1	Wait for all in-sync replicas to respond.
* `use_tls` - (Optional) Whether to use TLS for secure logging. Can be either true or false.
* `tls_ca_cert` - (Optional) A secure certificate to authenticate the server with. Must be in PEM format.
* `tls_client_cert` - (Optional) The client certificate used to make authenticated requests. Must be in PEM format.
* `tls_client_key` - (Optional) The client private key used to make authenticated requests. Must be in PEM format.
* `tls_hostname` - (Optional) The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN).
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `placement` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.
* `response_condition` - (Optional) The name of the `condition` to apply. If empty, always execute.

### logging_heroku block

The `logging_heroku` block supports:

* `name` - (Required) The unique name of the Heroku logging endpoint.
* `token` - (Required) The token to use for authentication (https://devcenter.heroku.com/articles/add-on-partner-log-integration).
* `url` - (Required) The url to stream logs to.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_honeycomb block

The `logging_honeycomb` block supports:

* `name` - (Required) The unique name of the Honeycomb logging endpoint.
* `dataset` - (Required) The Honeycomb Dataset you want to log to.
* `token` - (Required) The Write Key from the Account page of your Honeycomb account.
* `format` - (Optional) Apache style log formatting. Your log must produce valid JSON that Honeycomb can ingest.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_logshuttle block

The `logging_logshuttle` block supports:

* `name` - (Required) The unique name of the Logshuttle logging endpoint.
* `token` - (Required) The data authentication token associated with this endpoint.
* `url` - (Required) Your Log Shuttle endpoint url.
* `format` - (Optional) Apache style log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_openstack block

The `logging_openstack` block supports:

* `name` - (Required) The unique name of the OpenStack logging endpoint.
* `bucket_name` - (Required) The name of your OpenStack container.
* `url` - (Required) Your OpenStack auth url.
* `user` - (Required) The username for your OpenStack account.
* `access_key` - (Required) Your OpenStack account access key.
* `public_key` - (Optional) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `path` - (Optional) Path to store the files. Must end with a trailing slash.
If this field is left empty, the files will be saved in the bucket's root path.
* `period` - (Optional) How frequently the logs should be transferred, in
seconds. Default `3600`.
* `gzip_level` - (Optional) What level of Gzip encoding to have when dumping logs (default 0, no compression).
* `message_type` - (Optional) How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`. Default `classic`. [Fastly Documentation](https://developer.fastly.com/reference/api/logging/gcs/)
* `format` - (Optional) Apache style log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`. Default `2`.
* `timestamp_format` - (Optional) The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.
* `placement` - (Optional) Where in the generated VCL the logging call should be placed; one of: `none` or `waf_debug`.

### logging_digitalocean block

The `logging_digitalocean` block supports:

* `name` - (Required) The unique name of the DigitalOcean Spaces logging endpoint.
* `bucket_name` - (Required) The name of the DigitalOcean Space.
* `access_key` - (Required) Your DigitalOcean Spaces account access key.
* `secret_key` - (Required) Your DigitalOcean Spaces account secret key.
* `public_key` - (Optional) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `domain` - (Optional) The domain of the DigitalOcean Spaces endpoint (default "nyc3.digitaloceanspaces.com").
* `path` - (Optional) The path to upload logs to.
* `period` - (Optional) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
* `timestamp_format` - (Optional) The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `gzip_level` - (Optional) What level of Gzip encoding to have when dumping logs (default 0, no compression).
* `message_type` - (Optional) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
* `format` - (Optional) Apache-style string or VCL variables to use for log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_cloudfiles block

The `logging_cloudfiles` block supports:

* `name` - (Required) The unique name of the Rackspace Cloud Files logging endpoint.
* `user` - (Required) The username for your Cloud Files account.
* `bucket_name` - (Required) The name of your Cloud Files container.
* `access_key` - (Required) Your Cloud File account access key.
* `public_key` - (Optional) The PGP public key that Fastly will use to encrypt your log files before writing them to disk.
* `gzip_level` - (Optional) What level of GZIP encoding to have when dumping logs (default 0, no compression).
* `message_type` - (Optional) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
* `path` - (Optional) The path to upload logs to.
* `region` - (Optional) The region to stream logs to. One of: DFW (Dallas), ORD (Chicago), IAD (Northern Virginia), LON (London), SYD (Sydney), HKG (Hong Kong).
* `period` - (Optional) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
* `timestamp_format` - (Optional) The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).
* `format` - (Optional) Apache style log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### logging_kinesis block

The `logging_kinesis` block supports:

* `name` - (Required) The unique name of the Kinesis logging endpoint.
* `topic` - (Required) The Kinesis stream name.
* `region` - (Optional) The AWS region the stream resides in. (Default: `us-east-1`).
* `access_key` - (Required) The AWS access key to be used to write to the stream.
* `secret_key` - (Required) The AWS secret access key to authenticate with.
* `format` - (Optional) Apache style log formatting.
* `format_version` - (Optional) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
* `placement` - (Optional) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
* `response_condition` - (Optional) The name of an existing condition in the configured endpoint, or leave blank to always execute.

### response_object block

The `response_object` block supports:

* `name` - (Required) A unique name to identify this Response Object.
* `status` - (Optional) The HTTP Status Code. Default `200`.
* `response` - (Optional) The HTTP Response. Default `Ok`.
* `content` - (Optional) The content to deliver for the response object.
* `content_type` - (Optional) The MIME type of the content.
* `request_condition` - (Optional) Name of already defined `condition` to be checked during the request phase. If the condition passes then this object will be delivered. This `condition` must be of type `REQUEST`.
* `cache_condition` - (Optional) Name of already defined `condition` to check after we have retrieved an object. If the condition passes then deliver this Request Object instead. This `condition` must be of type `CACHE`. For detailed information about Conditionals,
see [Fastly's Documentation on Conditionals][fastly-conditionals].

### snippet block

The `snippet` block supports:

* `name` - (Required) A name that is unique across "regular" and "dynamic" VCL Snippet configuration blocks.
* `type` - (Required) The location in generated VCL where the snippet should be placed (can be one of `init`, `recv`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log` or `none`).
* `content` (Required) The VCL code that specifies exactly what the snippet does.
* `priority` - (Optional) Priority determines the ordering for multiple snippets. Lower numbers execute first.  Defaults to `100`.

### dynamicsnippet block

The `dynamicsnippet` block supports:

* `name` - (Required) A name that is unique across "regular" and "dynamic" VCL Snippet configuration blocks.
* `type` - (Required) The location in generated VCL where the snippet should be placed (can be one of `init`, `recv`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log` or `none`).
* `priority` - (Optional) Priority determines the ordering for multiple snippets. Lower numbers execute first.  Defaults to `100`.

### vcl block

The `vcl` block supports:

* `name` - (Required) A unique name for this configuration block.
* `content` - (Required) The custom VCL code to upload.
* `main` - (Optional) If `true`, use this block as the main configuration. If
`false`, use this block as an includable library. Only a single VCL block can be
marked as the main block. Default is `false`.

### acl block

The `acl` block supports:

* `name` - (Required) A unique name to identify this ACL.

### dictionary block

The `dictionary` block supports:

* `name` - (Required) A unique name to identify this dictionary.
* `write_only` - (Optional) If `true`, the dictionary is a private dictionary, and items are not readable in the UI or
via API. Default is `false`. It is important to note that changing this attribute will delete and recreate the
dictionary, discard the current items in the dictionary. Using a write-only/private dictionary should only be done if
the items are managed outside of Terraform.

### waf block

The `waf` block supports:

* `response_object` - (Required) The name of the [response object](#response_object) used by the Web Application Firewall.
* `prefetch_condition` - (Optional) The `condition` to determine which requests will be run past your Fastly WAF. This `condition` must be of type `PREFETCH`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals][fastly-conditionals].

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` – The ID of the Service.
* `active_version` – The currently active version of your Fastly Service.
* `cloned_version` - The latest cloned version by the provider. The value gets only set after running `terraform apply`.

The `dynamicsnippet` block exports:

* `snippet_id` - The ID of the dynamic snippet.

The `acl` block exports:

* `acl_id` - The ID of the ACL.

The `waf` block exports:

* `waf_id` - The ID of the WAF.

The `dictionary` block exports:

* `dictionary_id` - The ID of the dictionary.

[fastly-s3]: https://docs.fastly.com/en/guides/amazon-s3
[fastly-cname]: https://docs.fastly.com/en/guides/adding-cname-records
[fastly-conditionals]: https://docs.fastly.com/en/guides/using-conditions
[fastly-sumologic]: https://developer.fastly.com/reference/api/logging/sumologic/
[fastly-gcs]: https://developer.fastly.com/reference/api/logging/gcs/

## Import

Fastly Service can be imported using their service ID, e.g.

```
$ terraform import fastly_service_v1.demo xxxxxxxxxxxxxxxxxxxx
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **domain** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--domain))
- **name** (String) Unique name for this Service

### Optional

- **acl** (Block Set) (see [below for nested schema](#nestedblock--acl))
- **activate** (Boolean) Conditionally prevents the Service from being activated
- **backend** (Block Set) (see [below for nested schema](#nestedblock--backend))
- **bigquerylogging** (Block Set) (see [below for nested schema](#nestedblock--bigquerylogging))
- **blobstoragelogging** (Block Set) (see [below for nested schema](#nestedblock--blobstoragelogging))
- **cache_setting** (Block Set) (see [below for nested schema](#nestedblock--cache_setting))
- **comment** (String) A personal freeform descriptive note
- **condition** (Block Set) (see [below for nested schema](#nestedblock--condition))
- **default_host** (String) The default hostname for the version
- **default_ttl** (Number) The default Time-to-live (TTL) for the version
- **dictionary** (Block Set) (see [below for nested schema](#nestedblock--dictionary))
- **director** (Block Set) (see [below for nested schema](#nestedblock--director))
- **dynamicsnippet** (Block Set) (see [below for nested schema](#nestedblock--dynamicsnippet))
- **force_destroy** (Boolean)
- **gcslogging** (Block Set) (see [below for nested schema](#nestedblock--gcslogging))
- **gzip** (Block Set) (see [below for nested schema](#nestedblock--gzip))
- **header** (Block Set) (see [below for nested schema](#nestedblock--header))
- **healthcheck** (Block Set) (see [below for nested schema](#nestedblock--healthcheck))
- **httpslogging** (Block Set) (see [below for nested schema](#nestedblock--httpslogging))
- **id** (String) The ID of this resource.
- **logentries** (Block Set) (see [below for nested schema](#nestedblock--logentries))
- **logging_cloudfiles** (Block Set) (see [below for nested schema](#nestedblock--logging_cloudfiles))
- **logging_datadog** (Block Set) (see [below for nested schema](#nestedblock--logging_datadog))
- **logging_digitalocean** (Block Set) (see [below for nested schema](#nestedblock--logging_digitalocean))
- **logging_elasticsearch** (Block Set) (see [below for nested schema](#nestedblock--logging_elasticsearch))
- **logging_ftp** (Block Set) (see [below for nested schema](#nestedblock--logging_ftp))
- **logging_googlepubsub** (Block Set) (see [below for nested schema](#nestedblock--logging_googlepubsub))
- **logging_heroku** (Block Set) (see [below for nested schema](#nestedblock--logging_heroku))
- **logging_honeycomb** (Block Set) (see [below for nested schema](#nestedblock--logging_honeycomb))
- **logging_kafka** (Block Set) (see [below for nested schema](#nestedblock--logging_kafka))
- **logging_kinesis** (Block Set) (see [below for nested schema](#nestedblock--logging_kinesis))
- **logging_loggly** (Block Set) (see [below for nested schema](#nestedblock--logging_loggly))
- **logging_logshuttle** (Block Set) (see [below for nested schema](#nestedblock--logging_logshuttle))
- **logging_newrelic** (Block Set) (see [below for nested schema](#nestedblock--logging_newrelic))
- **logging_openstack** (Block Set) (see [below for nested schema](#nestedblock--logging_openstack))
- **logging_scalyr** (Block Set) (see [below for nested schema](#nestedblock--logging_scalyr))
- **logging_sftp** (Block Set) (see [below for nested schema](#nestedblock--logging_sftp))
- **papertrail** (Block Set) (see [below for nested schema](#nestedblock--papertrail))
- **request_setting** (Block Set) (see [below for nested schema](#nestedblock--request_setting))
- **response_object** (Block Set) (see [below for nested schema](#nestedblock--response_object))
- **s3logging** (Block Set) (see [below for nested schema](#nestedblock--s3logging))
- **snippet** (Block Set) (see [below for nested schema](#nestedblock--snippet))
- **splunk** (Block Set) (see [below for nested schema](#nestedblock--splunk))
- **sumologic** (Block Set) (see [below for nested schema](#nestedblock--sumologic))
- **syslog** (Block Set) (see [below for nested schema](#nestedblock--syslog))
- **vcl** (Block Set) (see [below for nested schema](#nestedblock--vcl))
- **version_comment** (String) A personal freeform descriptive note
- **waf** (Block List, Max: 1) (see [below for nested schema](#nestedblock--waf))

### Read-only

- **active_version** (Number)
- **cloned_version** (Number)

<a id="nestedblock--domain"></a>
### Nested Schema for `domain`

Required:

- **name** (String) The domain that this Service will respond to

Optional:

- **comment** (String)


<a id="nestedblock--acl"></a>
### Nested Schema for `acl`

Required:

- **name** (String) Unique name to refer to this ACL

Read-only:

- **acl_id** (String) Generated acl id


<a id="nestedblock--backend"></a>
### Nested Schema for `backend`

Required:

- **address** (String) An IPv4, hostname, or IPv6 address for the Backend
- **name** (String) A name for this Backend

Optional:

- **auto_loadbalance** (Boolean) Should this Backend be load balanced
- **between_bytes_timeout** (Number) How long to wait between bytes in milliseconds
- **connect_timeout** (Number) How long to wait for a timeout in milliseconds
- **error_threshold** (Number) Number of errors to allow before the Backend is marked as down
- **first_byte_timeout** (Number) How long to wait for the first bytes in milliseconds
- **healthcheck** (String) The healthcheck name that should be used for this Backend
- **max_conn** (Number) Maximum number of connections for this Backend
- **max_tls_version** (String) Maximum allowed TLS version on SSL connections to this backend.
- **min_tls_version** (String) Minimum allowed TLS version on SSL connections to this backend.
- **override_host** (String) The hostname to override the Host header
- **port** (Number) The port number Backend responds on. Default 80
- **request_condition** (String) Name of a condition, which if met, will select this backend during a request.
- **shield** (String) The POP of the shield designated to reduce inbound load.
- **ssl_ca_cert** (String) CA certificate attached to origin.
- **ssl_cert_hostname** (String) SSL certificate hostname for cert verification
- **ssl_check_cert** (Boolean) Be strict on checking SSL certs
- **ssl_ciphers** (String) Comma sepparated list of ciphers
- **ssl_client_cert** (String, Sensitive) SSL certificate file for client connections to the backend.
- **ssl_client_key** (String, Sensitive) SSL key file for client connections to backend.
- **ssl_hostname** (String) SSL certificate hostname
- **ssl_sni_hostname** (String) SSL certificate hostname for SNI verification
- **use_ssl** (Boolean) Whether or not to use SSL to reach the Backend
- **weight** (Number) The portion of traffic to send to a specific origins. Each origin receives weight/total of the traffic.


<a id="nestedblock--bigquerylogging"></a>
### Nested Schema for `bigquerylogging`

Required:

- **dataset** (String) The ID of your BigQuery dataset
- **name** (String) Unique name to refer to this logging setup
- **project_id** (String) The ID of your GCP project
- **table** (String) The ID of your BigQuery table

Optional:

- **email** (String, Sensitive) The email address associated with the target BigQuery dataset on your account.
- **format** (String) The logging format desired.
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) Name of a condition to apply this logging.
- **secret_key** (String, Sensitive) The secret key associated with the target BigQuery dataset on your account.
- **template** (String) Big query table name suffix template


<a id="nestedblock--blobstoragelogging"></a>
### Nested Schema for `blobstoragelogging`

Required:

- **account_name** (String) The unique Azure Blob Storage namespace in which your data objects are stored
- **container** (String) The name of the Azure Blob Storage container in which to store logs
- **name** (String) The unique name of the Azure Blob Storage logging endpoint

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t "%r" %>s %b`)
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)
- **gzip_level** (Number) The Gzip compression level (default: 0)
- **message_type** (String) How the message should be formatted (default: `classic`)
- **path** (String) The path to upload logs to. Must end with a trailing slash
- **period** (Number) How frequently the logs should be transferred, in seconds (default: 3600)
- **placement** (String) Where in the generated VCL the logging call should be placed
- **public_key** (String) The PGP public key that Fastly will use to encrypt your log files before writing them to disk
- **response_condition** (String) The name of the condition to apply
- **sas_token** (String, Sensitive) The Azure shared access signature providing write access to the blob service objects
- **timestamp_format** (String) strftime specified timestamp formatting (default: `%Y-%m-%dT%H:%M:%S.000`)


<a id="nestedblock--cache_setting"></a>
### Nested Schema for `cache_setting`

Required:

- **name** (String) A name to refer to this Cache Setting

Optional:

- **action** (String) Action to take
- **cache_condition** (String) Name of a condition to check if this Cache Setting applies
- **stale_ttl** (Number) Max 'Time To Live' for stale (unreachable) objects.
- **ttl** (Number) The 'Time To Live' for the object


<a id="nestedblock--condition"></a>
### Nested Schema for `condition`

Required:

- **name** (String)
- **statement** (String) The statement used to determine if the condition is met
- **type** (String) Type of the condition, either `REQUEST`, `RESPONSE`, or `CACHE`

Optional:

- **priority** (Number) A number used to determine the order in which multiple conditions execute. Lower numbers execute first


<a id="nestedblock--dictionary"></a>
### Nested Schema for `dictionary`

Required:

- **name** (String) Unique name to refer to this Dictionary

Optional:

- **write_only** (Boolean) Determines if items in the dictionary are readable or not

Read-only:

- **dictionary_id** (String) Generated dictionary ID


<a id="nestedblock--director"></a>
### Nested Schema for `director`

Required:

- **backends** (Set of String) List of backends associated with this director
- **name** (String) A name to refer to this director

Optional:

- **capacity** (Number) Load balancing weight for the backends
- **comment** (String)
- **quorum** (Number) Percentage of capacity that needs to be up for the director itself to be considered up
- **retries** (Number) How many backends to search if it fails
- **shield** (String) Selected POP to serve as a 'shield' for origin servers.
- **type** (Number) Type of load balance group to use. Integer, 1 to 4. Values: 1 (random), 3 (hash), 4 (client)


<a id="nestedblock--dynamicsnippet"></a>
### Nested Schema for `dynamicsnippet`

Required:

- **name** (String) A unique name to refer to this VCL snippet
- **type** (String) One of init, recv, hit, miss, pass, fetch, error, deliver, log, none

Optional:

- **priority** (Number) Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)

Read-only:

- **snippet_id** (String) Generated VCL snippet Id


<a id="nestedblock--gcslogging"></a>
### Nested Schema for `gcslogging`

Required:

- **bucket_name** (String) The name of the bucket in which to store the logs.
- **name** (String) Unique name to refer to this logging setup

Optional:

- **email** (String) The email address associated with the target GCS bucket on your account.
- **format** (String) Apache-style string or VCL variables to use for log formatting
- **gzip_level** (Number) Gzip Compression level
- **message_type** (String) The log message type per the fastly docs: https://docs.fastly.com/api/logging#logging_gcs
- **path** (String) Path to store the files. Must end with a trailing slash
- **period** (Number) How frequently the logs should be transferred, in seconds (Default 3600)
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) Name of a condition to apply this logging.
- **secret_key** (String, Sensitive) The secret key associated with the target gcs bucket on your account.
- **timestamp_format** (String) specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)


<a id="nestedblock--gzip"></a>
### Nested Schema for `gzip`

Required:

- **name** (String) A name to refer to this gzip condition

Optional:

- **cache_condition** (String) Name of a condition controlling when this gzip configuration applies.
- **content_types** (Set of String) Content types to apply automatic gzip to
- **extensions** (Set of String) File extensions to apply automatic gzip to. Do not include '.'


<a id="nestedblock--header"></a>
### Nested Schema for `header`

Required:

- **action** (String) One of set, append, delete, regex, or regex_repeat
- **destination** (String) Header this affects
- **name** (String) A name to refer to this Header object
- **type** (String) Type to manipulate: request, fetch, cache, response

Optional:

- **cache_condition** (String) Optional name of a cache condition to apply.
- **ignore_if_set** (Boolean) Don't add the header if it is already. (Only applies to 'set' action.). Default `false`
- **priority** (Number) Lower priorities execute first. (Default: 100.)
- **regex** (String) Regular expression to use (Only applies to 'regex' and 'regex_repeat' actions.)
- **request_condition** (String) Optional name of a request condition to apply.
- **response_condition** (String) Optional name of a response condition to apply.
- **source** (String) Variable to be used as a source for the header content (Does not apply to 'delete' action.)
- **substitution** (String) Value to substitute in place of regular expression. (Only applies to 'regex' and 'regex_repeat'.)


<a id="nestedblock--healthcheck"></a>
### Nested Schema for `healthcheck`

Required:

- **host** (String) Which host to check
- **name** (String) A name to refer to this healthcheck
- **path** (String) The path to check

Optional:

- **check_interval** (Number) How often to run the healthcheck in milliseconds
- **expected_response** (Number) The status code expected from the host
- **http_version** (String) Whether to use version 1.0 or 1.1 HTTP
- **initial** (Number) When loading a config, the initial number of probes to be seen as OK
- **method** (String) Which HTTP method to use
- **threshold** (Number) How many healthchecks must succeed to be considered healthy
- **timeout** (Number) Timeout in milliseconds
- **window** (Number) The number of most recent healthcheck queries to keep for this healthcheck


<a id="nestedblock--httpslogging"></a>
### Nested Schema for `httpslogging`

Required:

- **name** (String) The unique name of the HTTPS logging endpoint
- **url** (String) URL that log data will be sent to. Must use the https protocol.

Optional:

- **content_type** (String) Content-Type header sent with the request.
- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)
- **header_name** (String) Custom header sent with the request.
- **header_value** (String) Value of the custom header sent with the request.
- **json_format** (String) Formats log entries as JSON. Can be either disabled (`0`), array of json (`1`), or newline delimited json (`2`).
- **message_type** (String) How the message should be formatted
- **method** (String) HTTP method used for request.
- **placement** (String) Where in the generated VCL the logging call should be placed
- **request_max_bytes** (Number) The maximum number of bytes sent in one request.
- **request_max_entries** (Number) The maximum number of logs sent in one request.
- **response_condition** (String) The name of the condition to apply
- **tls_ca_cert** (String, Sensitive) A secure certificate to authenticate the server with. Must be in PEM format.
- **tls_client_cert** (String, Sensitive) The client certificate used to make authenticated requests. Must be in PEM format.
- **tls_client_key** (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format.
- **tls_hostname** (String) The hostname used to verify the server's certificate. It can either be the Common Name (CN) or blockAttributes Subject Alternative Name (SAN).


<a id="nestedblock--logentries"></a>
### Nested Schema for `logentries`

Required:

- **name** (String) Unique name to refer to this logging setup
- **token** (String) Use token based authentication (https://logentries.com/doc/input-token/)

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **port** (Number) The port number configured in Logentries
- **response_condition** (String) Name of blockAttributes condition to apply this logging.
- **use_tls** (Boolean) Whether to use TLS for secure logging


<a id="nestedblock--logging_cloudfiles"></a>
### Nested Schema for `logging_cloudfiles`

Required:

- **access_key** (String, Sensitive) Your Cloudfile account access key.
- **bucket_name** (String) The name of your Cloud Files container.
- **name** (String) The unique name of the Cloud Files logging endpoint.
- **user** (String) The username for your Cloudfile account.

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **gzip_level** (Number) What level of GZIP encoding to have when dumping logs (default 0, no compression).
- **message_type** (String) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
- **path** (String) The path to upload logs to.
- **period** (Number) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **public_key** (String) The PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **region** (String) The region to stream logs to. One of: DFW	(Dallas), ORD (Chicago), IAD (Northern Virginia), LON (London), SYD (Sydney), HKG (Hong Kong).
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- **timestamp_format** (String) The specified format of the log's timestamp (default `%Y-%m-%dT%H:%M:%S.000`).


<a id="nestedblock--logging_datadog"></a>
### Nested Schema for `logging_datadog`

Required:

- **name** (String) The unique name of the Datadog logging endpoint.
- **token** (String, Sensitive) The API key from your Datadog account.

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **region** (String) The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined.
- **response_condition** (String) The name of the condition to apply.


<a id="nestedblock--logging_digitalocean"></a>
### Nested Schema for `logging_digitalocean`

Required:

- **access_key** (String, Sensitive) Your DigitalOcean Spaces account access key.
- **bucket_name** (String) The name of the DigitalOcean Space.
- **name** (String) The unique name of the DigitalOcean Spaces logging endpoint.
- **secret_key** (String, Sensitive) Your DigitalOcean Spaces account secret key.

Optional:

- **domain** (String) The domain of the DigitalOcean Spaces endpoint (default: nyc3.digitaloceanspaces.com).
- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **gzip_level** (Number) What level of Gzip encoding to have when dumping logs (default 0, no compression).
- **message_type** (String) How the message should be formatted.
- **path** (String) The path to upload logs to.
- **period** (Number) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **public_key** (String) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- **timestamp_format** (String) strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).


<a id="nestedblock--logging_elasticsearch"></a>
### Nested Schema for `logging_elasticsearch`

Required:

- **index** (String) The name of the Elasticsearch index to send documents (logs) to.
- **name** (String) The unique name of the Elasticsearch logging endpoint.
- **url** (String) The Elasticsearch URL to stream logs to.

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **password** (String, Sensitive) BasicAuth password.
- **pipeline** (String) The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing.
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **request_max_bytes** (Number) The maximum number of bytes sent in one request.
- **request_max_entries** (Number) The maximum number of logs sent in one request.
- **response_condition** (String) The name of the condition to apply
- **tls_ca_cert** (String, Sensitive) A secure certificate to authenticate the server with. Must be in PEM format.
- **tls_client_cert** (String, Sensitive) The client certificate used to make authenticated requests. Must be in PEM format.
- **tls_client_key** (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format.
- **tls_hostname** (String) The hostname used to verify the server's certificate. It can either be the Common Name (CN) or blockAttributes Subject Alternative Name (SAN).
- **user** (String) BasicAuth user.


<a id="nestedblock--logging_ftp"></a>
### Nested Schema for `logging_ftp`

Required:

- **address** (String) The FTP URL to stream logs to.
- **name** (String) The unique name of the FTP logging endpoint.
- **password** (String, Sensitive) The password for the server (for anonymous use an email address).
- **path** (String) The path to upload log files to. If the path ends in / then it is treated as blockAttributes directory.
- **user** (String) The username for the server (can be anonymous).

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **gzip_level** (Number) Gzip Compression level.
- **message_type** (String) How the message should be formatted (default: `classic`)
- **period** (Number) How frequently the logs should be transferred, in seconds (Default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **port** (Number) The port number.
- **public_key** (String) The PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **response_condition** (String) The name of the condition to apply.
- **timestamp_format** (String) specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).


<a id="nestedblock--logging_googlepubsub"></a>
### Nested Schema for `logging_googlepubsub`

Required:

- **name** (String) The unique name of the Google Cloud Pub/Sub logging endpoint.
- **project_id** (String) The ID of your Google Cloud Platform project.
- **secret_key** (String) Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON.
- **topic** (String) The Google Cloud Pub/Sub topic to which logs will be published.
- **user** (String) Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON.

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_heroku"></a>
### Nested Schema for `logging_heroku`

Required:

- **name** (String) The unique name of the Heroku logging endpoint.
- **token** (String, Sensitive) The token to use for authentication (https://www.heroku.com/docs/customer-token-authentication-token/).
- **url** (String) The url to stream logs to.

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_honeycomb"></a>
### Nested Schema for `logging_honeycomb`

Required:

- **dataset** (String) The Honeycomb Dataset you want to log to.
- **name** (String) The unique name of the Honeycomb logging endpoint.
- **token** (String, Sensitive) The Write Key from the Account page of your Honeycomb account.

Optional:

- **format** (String) Apache style log formatting. Your log must produce valid JSON that Honeycomb can ingest.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_kafka"></a>
### Nested Schema for `logging_kafka`

Required:

- **brokers** (String) A comma-separated list of IP addresses or hostnames of Kafka brokers.
- **name** (String) The unique name of the Kafka logging endpoint.
- **topic** (String) The Kafka topic to send logs to.

Optional:

- **auth_method** (String) SASL authentication method. One of: plain, scram-sha-256, scram-sha-512.
- **compression_codec** (String) The codec used for compression of your logs. One of: gzip, snappy, lz4.
- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **parse_log_keyvals** (Boolean) Enables parsing of key=value tuples from the beginning of a logline, turning them into record headers.
- **password** (String) SASL Pass.
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **request_max_bytes** (Number) Maximum size of log batch, if non-zero. Defaults to 0 for unbounded.
- **required_acks** (String) The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0 No servers need to respond. -1	Wait for all in-sync replicas to respond.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- **tls_ca_cert** (String, Sensitive) A secure certificate to authenticate the server with. Must be in PEM format.
- **tls_client_cert** (String, Sensitive) The client certificate used to make authenticated requests. Must be in PEM format.
- **tls_client_key** (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format.
- **tls_hostname** (String) The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN).
- **use_tls** (Boolean) Whether to use TLS for secure logging. Can be either true or false.
- **user** (String) SASL User.


<a id="nestedblock--logging_kinesis"></a>
### Nested Schema for `logging_kinesis`

Required:

- **access_key** (String, Sensitive) The AWS access key to be used to write to the stream.
- **name** (String) The unique name of the Kinesis logging endpoint.
- **secret_key** (String, Sensitive) The AWS secret access key to authenticate with.
- **topic** (String) The Kinesis stream name.

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **region** (String) The AWS region the stream resides in.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_loggly"></a>
### Nested Schema for `logging_loggly`

Required:

- **name** (String) The unique name of the Loggly logging endpoint.
- **token** (String, Sensitive) The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/).

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_logshuttle"></a>
### Nested Schema for `logging_logshuttle`

Required:

- **name** (String) The unique name of the Log Shuttle logging endpoint.
- **token** (String, Sensitive) The data authentication token associated with this endpoint.
- **url** (String) Your Log Shuttle endpoint url.

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_newrelic"></a>
### Nested Schema for `logging_newrelic`

Required:

- **name** (String) The unique name of the New Relic logging endpoint
- **token** (String, Sensitive) The Insert API key from the Account page of your New Relic account.

Optional:

- **format** (String) Apache style log formatting. Your log must produce valid JSON that New Relic Logs can ingest.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) The name of the condition to apply.


<a id="nestedblock--logging_openstack"></a>
### Nested Schema for `logging_openstack`

Required:

- **access_key** (String, Sensitive) Your OpenStack account access key.
- **bucket_name** (String) The name of your OpenStack container.
- **name** (String) The unique name of the OpenStack logging endpoint.
- **url** (String) Your OpenStack auth url.
- **user** (String) The username for your OpenStack account.

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).
- **gzip_level** (Number) What level of Gzip encoding to have when dumping logs (default 0, no compression).
- **message_type** (String) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
- **path** (String) The path to upload logs to.
- **period** (Number) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.
- **public_key** (String) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.
- **timestamp_format** (String) specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)


<a id="nestedblock--logging_scalyr"></a>
### Nested Schema for `logging_scalyr`

Required:

- **name** (String) The unique name of the Scalyr logging endpoint.
- **token** (String, Sensitive) The token to use for authentication (https://www.scalyr.com/keys).

Optional:

- **format** (String) Apache style log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **region** (String) The region that log data will be sent to. One of US or EU. Defaults to US if undefined.
- **response_condition** (String) The name of an existing condition in the configured endpoint, or leave blank to always execute.


<a id="nestedblock--logging_sftp"></a>
### Nested Schema for `logging_sftp`

Required:

- **address** (String) The SFTP address to stream logs to.
- **name** (String) The unique name of the SFTP logging endpoint.
- **path** (String) The path to upload log files to. If the path ends in / then it is treated as blockAttributes directory.
- **ssh_known_hosts** (String) A list of host keys for all hosts we can connect to over SFTP.
- **user** (String) The username for the server.

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).
- **gzip_level** (Number) What level of GZIP encoding to have when dumping logs (default 0, no compression).
- **message_type** (String) How the message should be formatted. One of: classic (default), loggly, logplex or blank.
- **password** (String, Sensitive) The password for the server. If both password and secret_key are passed, secret_key will be preferred.
- **period** (Number) How frequently log files are finalized so they can be available for reading (in seconds, default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **port** (Number) The port the SFTP service listens on. (Default: 22).
- **public_key** (String) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **response_condition** (String) The name of the condition to apply.
- **secret_key** (String, Sensitive) The SSH private key for the server. If both password and secret_key are passed, secret_key will be preferred.
- **timestamp_format** (String) The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).


<a id="nestedblock--papertrail"></a>
### Nested Schema for `papertrail`

Required:

- **address** (String) The address of the papertrail service
- **name** (String) Unique name to refer to this logging setup
- **port** (Number) The port of the papertrail service

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) Name of blockAttributes condition to apply this logging


<a id="nestedblock--request_setting"></a>
### Nested Schema for `request_setting`

Required:

- **name** (String) Unique name to refer to this Request Setting

Optional:

- **action** (String) Allows you to terminate request handling and immediately perform an action
- **bypass_busy_wait** (Boolean) Disable collapsed forwarding
- **default_host** (String) the host header
- **force_miss** (Boolean) Force a cache miss for the request
- **force_ssl** (Boolean) Forces the request use SSL
- **geo_headers** (Boolean) Inject Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region
- **hash_keys** (String) Comma separated list of varnish request object fields that should be in the hash key
- **max_stale_age** (Number) How old an object is allowed to be, in seconds. Default `60`
- **request_condition** (String) Name of a request condition to apply. If there is no condition this setting will always be applied.
- **timer_support** (Boolean) Injects the X-Timer info into the request
- **xff** (String) X-Forwarded-For options


<a id="nestedblock--response_object"></a>
### Nested Schema for `response_object`

Required:

- **name** (String) Unique name to refer to this request object

Optional:

- **cache_condition** (String) Name of the condition checked after we have retrieved an object. If the condition passes then deliver this Request Object instead.
- **content** (String) The content to deliver for the response object
- **content_type** (String) The MIME type of the content
- **request_condition** (String) Name of the condition to be checked during the request phase to see if the object should be delivered
- **response** (String) The HTTP Response of the object
- **status** (Number) The HTTP Status Code of the object


<a id="nestedblock--s3logging"></a>
### Nested Schema for `s3logging`

Required:

- **bucket_name** (String) S3 Bucket name to store logs in.
- **name** (String) The unique name of the S3 logging endpoint.

Optional:

- **domain** (String) Bucket endpoint.
- **format** (String) Apache-style string or VCL variables to use for log formatting.
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1).
- **gzip_level** (Number) Gzip Compression level.
- **message_type** (String) How the message should be formatted.
- **path** (String) Path to store the files. Must end with blockAttributes trailing slash.
- **period** (Number) How frequently the logs should be transferred, in seconds (Default 3600).
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **public_key** (String) A PGP public key that Fastly will use to encrypt your log files before writing them to disk.
- **redundancy** (String) The S3 redundancy level.
- **response_condition** (String) Name of blockAttributes condition to apply this logging.
- **s3_access_key** (String, Sensitive) AWS Access Key.
- **s3_secret_key** (String, Sensitive) AWS Secret Key
- **server_side_encryption** (String) Specify what type of server side encryption should be used. Can be either `AES256` or `aws:kms`.
- **server_side_encryption_kms_key_id** (String) Optional server-side KMS Key Id. Must be set if server_side_encryption is set to `aws:kms`.
- **timestamp_format** (String) specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).


<a id="nestedblock--snippet"></a>
### Nested Schema for `snippet`

Required:

- **content** (String) The contents of the VCL snippet
- **name** (String) A unique name to refer to this VCL snippet
- **type** (String) One of init, recv, hit, miss, pass, fetch, error, deliver, log, none

Optional:

- **priority** (Number) Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)


<a id="nestedblock--splunk"></a>
### Nested Schema for `splunk`

Required:

- **name** (String) The unique name of the Splunk logging endpoint
- **url** (String) The Splunk URL to stream logs to

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t "%r" %>s %b`)
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)
- **placement** (String) Where in the generated VCL the logging call should be placed
- **response_condition** (String) The name of the condition to apply
- **tls_ca_cert** (String) A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SPLUNK_CA_CERT`.
- **tls_hostname** (String) The hostname used to verify the server's certificate. It can either be the Common Name or blockAttributes Subject Alternative Name (SAN).
- **token** (String, Sensitive) The Splunk token to be used for authentication


<a id="nestedblock--sumologic"></a>
### Nested Schema for `sumologic`

Required:

- **name** (String) Unique name to refer to this logging setup
- **url** (String) The URL to POST to.

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting
- **format_version** (Number) The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)
- **message_type** (String) How the message should be formatted.
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **response_condition** (String) Name of blockAttributes condition to apply this logging.


<a id="nestedblock--syslog"></a>
### Nested Schema for `syslog`

Required:

- **address** (String) The address of the syslog service
- **name** (String) Unique name to refer to this logging setup

Optional:

- **format** (String) Apache-style string or VCL variables to use for log formatting
- **format_version** (Number) The version of the custom logging format. Can be either 1 or 2. (Default: 1)
- **message_type** (String) How the message should be formatted.
- **placement** (String) Where in the generated VCL the logging call should be placed.
- **port** (Number) The port of the syslog service
- **response_condition** (String) Name of blockAttributes condition to apply this logging.
- **tls_ca_cert** (String) A secure certificate to authenticate the server with. Must be in PEM format.
- **tls_client_cert** (String) The client certificate used to make authenticated requests. Must be in PEM format.
- **tls_client_key** (String, Sensitive) The client private key used to make authenticated requests. Must be in PEM format.
- **tls_hostname** (String) Used during the TLS handshake to validate the certificate.
- **token** (String) Authentication token
- **use_tls** (Boolean) Use TLS for secure logging


<a id="nestedblock--vcl"></a>
### Nested Schema for `vcl`

Required:

- **content** (String) The contents of this VCL configuration
- **name** (String) A name to refer to this VCL configuration

Optional:

- **main** (Boolean) Should this VCL configuration be the main configuration


<a id="nestedblock--waf"></a>
### Nested Schema for `waf`

Required:

- **response_object** (String) The Web Application Firewall's (WAF) response object

Optional:

- **disabled** (Boolean) A flag used to completely disable a Web Application Firewall. This is intended to only be used in an emergency.
- **prefetch_condition** (String) The Web Application Firewall's (WAF) prefetch condition

Read-only:

- **waf_id** (String) The Web Application Firewall (WAF) ID
