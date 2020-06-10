package fastly

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var fastlyNoServiceFoundErr = errors.New("No matching Fastly Service found")

func resourceServiceV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceV1Create,
		Read:   resourceServiceV1Read,
		Update: resourceServiceV1Update,
		Delete: resourceServiceV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for this Service",
			},

			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Managed by Terraform",
				Description: "A personal freeform descriptive note",
			},

			"version_comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A personal freeform descriptive note",
			},

			// Active Version represents the currently activated version in Fastly. In
			// Terraform, we abstract this number away from the users and manage
			// creating and activating. It's used internally, but also exported for
			// users to see.
			"active_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Cloned Version represents the latest cloned version by the provider. It
			// gets set whenever Terraform detects changes and clones the currently
			// activated version in order to modify it. Active Version and Cloned
			// Version can be different if the Activate field is set to false in order
			// to prevent the service from being activated. It is not used internally,
			// but it is exported for users to see after running `terraform apply`.
			"cloned_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"activate": {
				Type:        schema.TypeBool,
				Description: "Conditionally prevents the Service from being activated",
				Default:     true,
				Optional:    true,
			},

			"domain": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The domain that this Service will respond to",
						},

						"comment": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"condition": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"statement": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The statement used to determine if the condition is met",
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Type of the condition, either `REQUEST`, `RESPONSE`, or `CACHE`",
							ValidateFunc: validateConditionType(),
						},
					},
				},
			},

			"default_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
				Description: "The default Time-to-live (TTL) for the version",
			},

			"default_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The default hostname for the version",
			},

			"healthcheck": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this healthcheck",
						},
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Which host to check",
						},
						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The path to check",
						},
						// optional fields
						"check_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     5000,
							Description: "How often to run the healthcheck in milliseconds",
						},
						"expected_response": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "The status code expected from the host",
						},
						"http_version": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "1.1",
							Description: "Whether to use version 1.0 or 1.1 HTTP",
						},
						"initial": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     2,
							Description: "When loading a config, the initial number of probes to be seen as OK",
						},
						"method": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "HEAD",
							Description: "Which HTTP method to use",
						},
						"threshold": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3,
							Description: "How many healthchecks must succeed to be considered healthy",
						},
						"timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     500,
							Description: "Timeout in milliseconds",
						},
						"window": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     5,
							Description: "The number of most recent healthcheck queries to keep for this healthcheck",
						},
					},
				},
			},

			"backend": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name for this Backend",
						},
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "An IPv4, hostname, or IPv6 address for the Backend",
						},
						// Optional fields, defaults where they exist
						"auto_loadbalance": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Should this Backend be load balanced",
						},
						"between_bytes_timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10000,
							Description: "How long to wait between bytes in milliseconds",
						},
						"connect_timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1000,
							Description: "How long to wait for a timeout in milliseconds",
						},
						"error_threshold": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Number of errors to allow before the Backend is marked as down",
						},
						"first_byte_timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     15000,
							Description: "How long to wait for the first bytes in milliseconds",
						},
						"healthcheck": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The healthcheck name that should be used for this Backend",
						},
						"max_conn": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximum number of connections for this Backend",
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     80,
							Description: "The port number Backend responds on. Default 80",
						},
						"override_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The hostname to override the Host header",
						},
						"request_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition, which if met, will select this backend during a request.",
						},
						"shield": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The POP of the shield designated to reduce inbound load.",
						},
						"use_ssl": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not to use SSL to reach the Backend",
						},
						"max_tls_version": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Maximum allowed TLS version on SSL connections to this backend.",
						},
						"min_tls_version": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Minimum allowed TLS version on SSL connections to this backend.",
						},
						"ssl_ciphers": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Comma sepparated list of ciphers",
						},
						"ssl_check_cert": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Be strict on checking SSL certs",
						},
						"ssl_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSL certificate hostname",
							Deprecated:  "Use ssl_cert_hostname and ssl_sni_hostname instead.",
						},
						"ssl_ca_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "CA certificate attached to origin.",
						},
						"ssl_cert_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSL certificate hostname for cert verification",
						},
						"ssl_sni_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSL certificate hostname for SNI verification",
						},
						"ssl_client_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSL certificate file for client connections to the backend.",
							Sensitive:   true,
						},
						"ssl_client_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSL key file for client connections to backend.",
							Sensitive:   true,
						},

						"weight": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "The portion of traffic to send to a specific origins. Each origin receives weight/total of the traffic.",
						},
					},
				},
			},

			"director": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this director",
						},
						"backends": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "List of backends associated with this director",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						// optional fields
						"capacity": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Load balancing weight for the backends",
						},
						"comment": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"shield": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Selected POP to serve as a 'shield' for origin servers.",
						},
						"quorum": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      75,
							Description:  "Percentage of capacity that needs to be up for the director itself to be considered up",
							ValidateFunc: validateDirectorQuorum(),
						},
						"type": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							Description:  "Type of load balance group to use. Integer, 1 to 4. Values: 1 (random), 3 (hash), 4 (client)",
							ValidateFunc: validateDirectorType(),
						},
						"retries": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     5,
							Description: "How many backends to search if it fails",
						},
					},
				},
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"cache_setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this Cache Setting",
						},
						"action": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Action to take",
						},
						// optional
						"cache_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to check if this Cache Setting applies",
						},
						"stale_ttl": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max 'Time To Live' for stale (unreachable) objects.",
						},
						"ttl": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The 'Time To Live' for the object",
						},
					},
				},
			},

			"gzip": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this gzip condition",
						},
						// optional fields
						"content_types": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Content types to apply automatic gzip to",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"extensions": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "File extensions to apply automatic gzip to. Do not include '.'",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"cache_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition controlling when this gzip configuration applies.",
						},
					},
				},
			},

			"header": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this Header object",
						},
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "One of set, append, delete, regex, or regex_repeat",
							ValidateFunc: validateHeaderAction(),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Type to manipulate: request, fetch, cache, response",
							ValidateFunc: validateHeaderType(),
						},
						"destination": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Header this affects",
						},
						// Optional fields, defaults where they exist
						"ignore_if_set": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Don't add the header if it is already. (Only applies to 'set' action.). Default `false`",
						},
						"source": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Variable to be used as a source for the header content (Does not apply to 'delete' action.)",
						},
						"regex": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Regular expression to use (Only applies to 'regex' and 'regex_repeat' actions.)",
						},
						"substitution": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Value to substitute in place of regular expression. (Only applies to 'regex' and 'regex_repeat'.)",
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Lower priorities execute first. (Default: 100.)",
						},
						"request_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Optional name of a request condition to apply.",
						},
						"cache_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Optional name of a cache condition to apply.",
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Optional name of a response condition to apply.",
						},
					},
				},
			},

			"s3logging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"bucket_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "S3 Bucket name to store logs in",
						},
						"s3_access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_ACCESS_KEY", ""),
							Description: "AWS Access Key",
							Sensitive:   true,
						},
						"s3_secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_SECRET_KEY", ""),
							Description: "AWS Secret Key",
							Sensitive:   true,
						},
						// Optional fields
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Path to store the files. Must end with a trailing slash",
						},
						"domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Bucket endpoint",
							Default:     "s3.amazonaws.com",
						},
						"gzip_level": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Gzip Compression level",
						},
						"period": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3600,
							Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"timestamp_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%Y-%m-%dT%H:%M:%S.000",
							Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
						},
						"redundancy": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The S3 redundancy level.",
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"message_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "classic",
							Description:  "How the message should be formatted.",
							ValidateFunc: validateLoggingMessageType(),
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
						"server_side_encryption": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Specify what type of server side encryption should be used. Can be either `AES256` or `aws:kms`.",
							ValidateFunc: validateLoggingServerSideEncryption(),
						},
						"server_side_encryption_kms_key_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional server-side KMS Key Id. Must be set if server_side_encryption is set to `aws:kms`",
						},
					},
				},
			},

			"papertrail": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The address of the papertrail service",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The port of the papertrail service",
						},
						// Optional fields
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging",
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"sumologic": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The URL to POST to.",
						},
						// Optional fields
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"message_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "classic",
							Description:  "How the message should be formatted.",
							ValidateFunc: validateLoggingMessageType(),
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"gcslogging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_EMAIL", ""),
							Description: "The email address associated with the target GCS bucket on your account.",
						},
						"bucket_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the bucket in which to store the logs.",
						},
						"secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_SECRET_KEY", ""),
							Description: "The secret key associated with the target gcs bucket on your account.",
							Sensitive:   true,
						},
						// Optional fields
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Path to store the files. Must end with a trailing slash",
						},
						"gzip_level": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Gzip Compression level",
						},
						"period": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3600,
							Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"timestamp_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%Y-%m-%dT%H:%M:%S.000",
							Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"message_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "classic",
							Description: "The log message type per the fastly docs: https://docs.fastly.com/api/logging#logging_gcs",
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"bigquerylogging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of your GCP project",
						},
						"dataset": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of your BigQuery dataset",
						},
						"table": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of your BigQuery table",
						},
						// Optional fields
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_BQ_EMAIL", ""),
							Description: "The email address associated with the target BigQuery dataset on your account.",
							Sensitive:   true,
						},
						"secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_BQ_SECRET_KEY", ""),
							Description: "The secret key associated with the target BigQuery dataset on your account.",
							Sensitive:   true,
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The logging format desired.",
							Default:     "%h %l %u %t \"%r\" %>s %b",
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"template": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Big query table name suffix template",
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"syslog": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The address of the syslog service",
						},
						// Optional
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     514,
							Description: "The port of the syslog service",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t \"%r\" %>s %b",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							Description:  "The version of the custom logging format. Can be either 1 or 2. (Default: 1)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"token": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Authentication token",
						},
						"use_tls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use TLS for secure logging",
						},
						"tls_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Used during the TLS handshake to validate the certificate.",
						},
						"tls_ca_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CA_CERT", ""),
							Description: "A secure certificate to authenticate the server with. Must be in PEM format.",
						},
						"tls_client_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_CERT", ""),
							Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
						},
						"tls_client_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_KEY", ""),
							Description: "The client private key used to make authenticated requests. Must be in PEM format.",
							Sensitive:   true,
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"message_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "classic",
							Description:  "How the message should be formatted.",
							ValidateFunc: validateLoggingMessageType(),
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"logentries": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"token": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Use token based authentication (https://logentries.com/doc/input-token/)",
						},
						// Optional
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     20000,
							Description: "The port number configured in Logentries",
						},
						"use_tls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to use TLS for secure logging",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a condition to apply this logging.",
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed.",
							ValidateFunc: validateLoggingPlacement(),
						},
					},
				},
			},

			"splunk": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique name of the Splunk logging endpoint",
						},
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Splunk URL to stream logs to",
						},
						"token": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_TOKEN", ""),
							Description: "The Splunk token to be used for authentication",
							Sensitive:   true,
						},
						// Optional fields
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t \"%r\" %>s %b",
							Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed",
							ValidateFunc: validateLoggingPlacement(),
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the condition to apply",
						},
						"tls_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN).",
						},
						"tls_ca_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_CA_CERT", ""),
							Description: "A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SPLUNK_CA_CERT`.",
						},
					},
				},
			},

			"blobstoragelogging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique name of the Azure Blob Storage logging endpoint",
						},
						"account_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique Azure Blob Storage namespace in which your data objects are stored",
						},
						"container": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the Azure Blob Storage container in which to store logs",
						},
						"sas_token": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", ""),
							Description: "The Azure shared access signature providing write access to the blob service objects",
							Sensitive:   true,
						},
						// Optional fields
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The path to upload logs to. Must end with a trailing slash",
						},
						"period": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3600,
							Description: "How frequently the logs should be transferred, in seconds (default: 3600)",
						},
						"timestamp_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%Y-%m-%dT%H:%M:%S.000",
							Description: "strftime specified timestamp formatting (default: `%Y-%m-%dT%H:%M:%S.000`)",
						},
						"gzip_level": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The Gzip compression level (default: 0)",
						},
						"public_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The PGP public key that Fastly will use to encrypt your log files before writing them to disk",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t \"%r\" %>s %b",
							Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
						},
						"format_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
							ValidateFunc: validateLoggingFormatVersion(),
						},
						"message_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "classic",
							Description:  "How the message should be formatted (default: `classic`)",
							ValidateFunc: validateLoggingMessageType(),
						},
						"placement": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Where in the generated VCL the logging call should be placed",
							ValidateFunc: validateLoggingPlacement(),
						},
						"response_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the condition to apply",
						},
					},
				},
			},

			"httpslogging": httpsloggingSchema,
			"response_object": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this request object",
						},
						// Optional fields
						"status": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "The HTTP Status Code of the object",
						},
						"response": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "OK",
							Description: "The HTTP Response of the object",
						},
						"content": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The content to deliver for the response object",
						},
						"content_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The MIME type of the content",
						},
						"request_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of the condition to be checked during the request phase to see if the object should be delivered",
						},
						"cache_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of the condition checked after we have retrieved an object. If the condition passes then deliver this Request Object instead.",
						},
					},
				},
			},

			"request_setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this Request Setting",
						},
						// Optional fields
						"request_condition": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Name of a request condition to apply. If there is no condition this setting will always be applied.",
						},
						"max_stale_age": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "How old an object is allowed to be, in seconds. Default `60`",
						},
						"force_miss": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Force a cache miss for the request",
						},
						"force_ssl": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Forces the request use SSL",
						},
						"action": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Allows you to terminate request handling and immediately perform an action",
						},
						"bypass_busy_wait": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Disable collapsed forwarding",
						},
						"hash_keys": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Comma separated list of varnish request object fields that should be in the hash key",
						},
						"xff": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "append",
							Description: "X-Forwarded-For options",
						},
						"timer_support": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Injects the X-Timer info into the request",
						},
						"geo_headers": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Inject Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region",
						},
						"default_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "the host header",
						},
					},
				},
			},

			"vcl": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this VCL configuration",
						},
						"content": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The contents of this VCL configuration",
						},
						"main": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Should this VCL configuration be the main configuration",
						},
					},
				},
			},

			"snippet": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A unique name to refer to this VCL snippet",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "One of init, recv, hit, miss, pass, fetch, error, deliver, log, none",
							ValidateFunc: validateSnippetType(),
						},
						"content": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The contents of the VCL snippet",
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)",
						},
					},
				},
			},
			"dynamicsnippet": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A unique name to refer to this VCL snippet",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "One of init, recv, hit, miss, pass, fetch, error, deliver, log, none",
							ValidateFunc: validateSnippetType(),
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)",
						},
						"snippet_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Generated VCL snippet Id",
						},
					},
				},
			},
			"acl": aclSchema,
			"dictionary": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this Dictionary",
						},
						// Optional fields
						"dictionary_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Generated dictionary ID",
						},
						"write_only": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Determines if items in the dictionary are readable or not",
						},
					},
				},
			},
		},
	}
}

func resourceServiceV1Create(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn
	service, err := conn.CreateService(&gofastly.CreateServiceInput{
		Name:    d.Get("name").(string),
		Comment: d.Get("comment").(string),
	})

	if err != nil {
		return err
	}

	d.SetId(service.ID)
	return resourceServiceV1Update(d, meta)
}

func resourceServiceV1Update(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn

	// Update Name and/or Comment. No new verions is required for this
	if d.HasChange("name") || d.HasChange("comment") {
		_, err := conn.UpdateService(&gofastly.UpdateServiceInput{
			ID:      d.Id(),
			Name:    d.Get("name").(string),
			Comment: d.Get("comment").(string),
		})
		if err != nil {
			return err
		}
	}

	// Once activated, Versions are locked and become immutable. This is true for
	// versions that are no longer active. For Domains, Backends, DefaultHost and
	// DefaultTTL, a new Version must be created first, and updates posted to that
	// Version. Loop these attributes and determine if we need to create a new version first
	var needsChange bool
	for _, v := range []string{
		"domain",
		"backend",
		"default_host",
		"default_ttl",
		"director",
		"header",
		"gzip",
		"healthcheck",
		"s3logging",
		"papertrail",
		"gcslogging",
		"bigquerylogging",
		"syslog",
		"sumologic",
		"logentries",
		"splunk",
		"blobstoragelogging",
		"httpslogging",
		"response_object",
		"condition",
		"request_setting",
		"cache_setting",
		"snippet",
		"dynamicsnippet",
		"vcl",
		"acl",
		"dictionary",
	} {
		if d.HasChange(v) {
			needsChange = true
		}
	}

	// Update the active version's comment. No new version is required for this
	if d.HasChange("version_comment") && !needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated
			latestVersion = 1
		}

		opts := gofastly.UpdateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
			Comment: d.Get("version_comment").(string),
		}

		log.Printf("[DEBUG] Update Version opts: %#v", opts)
		_, err := conn.UpdateVersion(&opts)
		if err != nil {
			return err
		}
	}

	initialVersion := false

	if needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			initialVersion = true
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated
			latestVersion = 1
		} else {
			// Clone the latest version, giving us an unlocked version we can modify
			log.Printf("[DEBUG] Creating clone of version (%d) for updates", latestVersion)
			newVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return err
			}

			// The new version number is named "Number", but it's actually a string
			latestVersion = newVersion.Number
			d.Set("cloned_version", latestVersion)

			// New versions are not immediately found in the API, or are not
			// immediately mutable, so we need to sleep a few and let Fastly ready
			// itself. Typically, 7 seconds is enough
			log.Print("[DEBUG] Sleeping 7 seconds to allow Fastly Version to be available")
			time.Sleep(7 * time.Second)

			// Update the cloned version's comment
			if d.Get("version_comment").(string) != "" {
				opts := gofastly.UpdateVersionInput{
					Service: d.Id(),
					Version: latestVersion,
					Comment: d.Get("version_comment").(string),
				}

				log.Printf("[DEBUG] Update Version opts: %#v", opts)
				_, err := conn.UpdateVersion(&opts)
				if err != nil {
					return err
				}
			}
		}

		// update general settings

		// If the requested default_ttl is 0, and this is the first
		// version being created, HasChange will return false, but we need
		// to set it anyway, so ensure we update the settings in that
		// case.
		if d.HasChange("default_host") || d.HasChange("default_ttl") || (d.Get("default_ttl") == 0 && initialVersion) {
			opts := gofastly.UpdateSettingsInput{
				Service: d.Id(),
				Version: latestVersion,
				// default_ttl has the same default value of 3600 that is provided by
				// the Fastly API, so it's safe to include here
				DefaultTTL: uint(d.Get("default_ttl").(int)),
			}

			if attr, ok := d.GetOk("default_host"); ok {
				opts.DefaultHost = attr.(string)
			}

			log.Printf("[DEBUG] Update Settings opts: %#v", opts)
			_, err := conn.UpdateSettings(&opts)
			if err != nil {
				return err
			}
		}

		// Conditions need to be updated first, as they can be referenced by other
		// configuraiton objects (Backends, Request Headers, etc)

		// Find difference in Conditions
		if d.HasChange("condition") {
			// Note: we don't utilize the PUT endpoint to update these objects, we simply
			// destroy any that have changed, and create new ones with the updated
			// values. This is how Terraform works with nested sub resources, we only
			// get the full diff not a partial set item diff. Because this is done
			// on a new version of the Fastly Service configuration, this is considered safe

			oc, nc := d.GetChange("condition")
			if oc == nil {
				oc = new(schema.Set)
			}
			if nc == nil {
				nc = new(schema.Set)
			}

			ocs := oc.(*schema.Set)
			ncs := nc.(*schema.Set)
			removeConditions := ocs.Difference(ncs).List()
			addConditions := ncs.Difference(ocs).List()

			// DELETE old Conditions
			for _, cRaw := range removeConditions {
				cf := cRaw.(map[string]interface{})
				opts := gofastly.DeleteConditionInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    cf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Conditions Removal opts: %#v", opts)
				err := conn.DeleteCondition(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new Conditions
			for _, cRaw := range addConditions {
				cf := cRaw.(map[string]interface{})
				opts := gofastly.CreateConditionInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    cf["name"].(string),
					Type:    cf["type"].(string),
					// need to trim leading/tailing spaces, incase the config has HEREDOC
					// formatting and contains a trailing new line
					Statement: strings.TrimSpace(cf["statement"].(string)),
					Priority:  cf["priority"].(int),
				}

				log.Printf("[DEBUG] Create Conditions Opts: %#v", opts)
				_, err := conn.CreateCondition(&opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in domains
		if d.HasChange("domain") {
			od, nd := d.GetChange("domain")
			if od == nil {
				od = new(schema.Set)
			}
			if nd == nil {
				nd = new(schema.Set)
			}

			ods := od.(*schema.Set)
			nds := nd.(*schema.Set)

			remove := ods.Difference(nds).List()
			add := nds.Difference(ods).List()

			// Delete removed domains
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteDomainInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Domain removal opts: %#v", opts)
				err := conn.DeleteDomain(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new Domains
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateDomainInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				if v, ok := df["comment"]; ok {
					opts.Comment = v.(string)
				}

				log.Printf("[DEBUG] Fastly Domain Addition opts: %#v", opts)
				_, err := conn.CreateDomain(&opts)
				if err != nil {
					return err
				}
			}
		}

		// Healthchecks need to be updated BEFORE backends
		if d.HasChange("healthcheck") {
			oh, nh := d.GetChange("healthcheck")
			if oh == nil {
				oh = new(schema.Set)
			}
			if nh == nil {
				nh = new(schema.Set)
			}

			ohs := oh.(*schema.Set)
			nhs := nh.(*schema.Set)
			removeHealthCheck := ohs.Difference(nhs).List()
			addHealthCheck := nhs.Difference(ohs).List()

			// DELETE old healthcheck configurations
			for _, hRaw := range removeHealthCheck {
				hf := hRaw.(map[string]interface{})
				opts := gofastly.DeleteHealthCheckInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    hf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Healthcheck removal opts: %#v", opts)
				err := conn.DeleteHealthCheck(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Healthcheck
			for _, hRaw := range addHealthCheck {
				hf := hRaw.(map[string]interface{})

				opts := gofastly.CreateHealthCheckInput{
					Service:          d.Id(),
					Version:          latestVersion,
					Name:             hf["name"].(string),
					Host:             hf["host"].(string),
					Path:             hf["path"].(string),
					CheckInterval:    uint(hf["check_interval"].(int)),
					ExpectedResponse: uint(hf["expected_response"].(int)),
					HTTPVersion:      hf["http_version"].(string),
					Initial:          uint(hf["initial"].(int)),
					Method:           hf["method"].(string),
					Threshold:        uint(hf["threshold"].(int)),
					Timeout:          uint(hf["timeout"].(int)),
					Window:           uint(hf["window"].(int)),
				}

				log.Printf("[DEBUG] Create Healthcheck Opts: %#v", opts)
				_, err := conn.CreateHealthCheck(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in backends
		if d.HasChange("backend") {
			ob, nb := d.GetChange("backend")
			if ob == nil {
				ob = new(schema.Set)
			}
			if nb == nil {
				nb = new(schema.Set)
			}

			obs := ob.(*schema.Set)
			nbs := nb.(*schema.Set)
			removeBackends := obs.Difference(nbs).List()
			addBackends := nbs.Difference(obs).List()

			// DELETE old Backends
			for _, bRaw := range removeBackends {
				bf := bRaw.(map[string]interface{})
				opts := gofastly.DeleteBackendInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    bf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Backend removal opts: %#v", opts)
				err := conn.DeleteBackend(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// Find and post new Backends
			for _, dRaw := range addBackends {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateBackendInput{
					Service:             d.Id(),
					Version:             latestVersion,
					Name:                df["name"].(string),
					Address:             df["address"].(string),
					OverrideHost:        df["override_host"].(string),
					AutoLoadbalance:     gofastly.CBool(df["auto_loadbalance"].(bool)),
					SSLCheckCert:        gofastly.CBool(df["ssl_check_cert"].(bool)),
					SSLHostname:         df["ssl_hostname"].(string),
					SSLCACert:           df["ssl_ca_cert"].(string),
					SSLCertHostname:     df["ssl_cert_hostname"].(string),
					SSLSNIHostname:      df["ssl_sni_hostname"].(string),
					UseSSL:              gofastly.CBool(df["use_ssl"].(bool)),
					SSLClientKey:        df["ssl_client_key"].(string),
					SSLClientCert:       df["ssl_client_cert"].(string),
					MaxTLSVersion:       df["max_tls_version"].(string),
					MinTLSVersion:       df["min_tls_version"].(string),
					SSLCiphers:          strings.Split(df["ssl_ciphers"].(string), ","),
					Shield:              df["shield"].(string),
					Port:                uint(df["port"].(int)),
					BetweenBytesTimeout: uint(df["between_bytes_timeout"].(int)),
					ConnectTimeout:      uint(df["connect_timeout"].(int)),
					ErrorThreshold:      uint(df["error_threshold"].(int)),
					FirstByteTimeout:    uint(df["first_byte_timeout"].(int)),
					MaxConn:             uint(df["max_conn"].(int)),
					Weight:              uint(df["weight"].(int)),
					RequestCondition:    df["request_condition"].(string),
					HealthCheck:         df["healthcheck"].(string),
				}

				log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
				_, err := conn.CreateBackend(&opts)
				if err != nil {
					return err
				}
			}
		}

		if d.HasChange("director") {
			od, nd := d.GetChange("director")
			if od == nil {
				od = new(schema.Set)
			}
			if nd == nil {
				nd = new(schema.Set)
			}

			ods := od.(*schema.Set)
			nds := nd.(*schema.Set)

			removeDirector := ods.Difference(nds).List()
			addDirector := nds.Difference(ods).List()

			// DELETE old director configurations
			for _, dRaw := range removeDirector {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteDirectorInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Director Removal opts: %#v", opts)
				err := conn.DeleteDirector(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Director
			for _, dRaw := range addDirector {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateDirectorInput{
					Service:  d.Id(),
					Version:  latestVersion,
					Name:     df["name"].(string),
					Comment:  df["comment"].(string),
					Shield:   df["shield"].(string),
					Capacity: uint(df["capacity"].(int)),
					Quorum:   uint(df["quorum"].(int)),
					Retries:  uint(df["retries"].(int)),
				}

				switch df["type"].(int) {
				case 1:
					opts.Type = gofastly.DirectorTypeRandom
				case 2:
					opts.Type = gofastly.DirectorTypeRoundRobin
				case 3:
					opts.Type = gofastly.DirectorTypeHash
				case 4:
					opts.Type = gofastly.DirectorTypeClient
				}

				log.Printf("[DEBUG] Director Create opts: %#v", opts)
				_, err := conn.CreateDirector(&opts)
				if err != nil {
					return err
				}

				if v, ok := df["backends"]; ok {
					if len(v.(*schema.Set).List()) > 0 {
						for _, b := range v.(*schema.Set).List() {
							opts := gofastly.CreateDirectorBackendInput{
								Service:  d.Id(),
								Version:  latestVersion,
								Director: df["name"].(string),
								Backend:  b.(string),
							}

							log.Printf("[DEBUG] Director Backend Create opts: %#v", opts)
							_, err := conn.CreateDirectorBackend(&opts)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

		if d.HasChange("header") {
			oh, nh := d.GetChange("header")
			if oh == nil {
				oh = new(schema.Set)
			}
			if nh == nil {
				nh = new(schema.Set)
			}

			ohs := oh.(*schema.Set)
			nhs := nh.(*schema.Set)

			remove := ohs.Difference(nhs).List()
			add := nhs.Difference(ohs).List()

			// Delete removed headers
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteHeaderInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Header removal opts: %#v", opts)
				err := conn.DeleteHeader(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new Headers
			for _, dRaw := range add {
				opts, err := buildHeader(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Header: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly Header Addition opts: %#v", opts)
				_, err = conn.CreateHeader(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in Gzips
		if d.HasChange("gzip") {
			og, ng := d.GetChange("gzip")
			if og == nil {
				og = new(schema.Set)
			}
			if ng == nil {
				ng = new(schema.Set)
			}

			ogs := og.(*schema.Set)
			ngs := ng.(*schema.Set)

			remove := ogs.Difference(ngs).List()
			add := ngs.Difference(ogs).List()

			// Delete removed gzip rules
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteGzipInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Gzip removal opts: %#v", opts)
				err := conn.DeleteGzip(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new Gzips
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateGzipInput{
					Service:        d.Id(),
					Version:        latestVersion,
					Name:           df["name"].(string),
					CacheCondition: df["cache_condition"].(string),
				}

				if v, ok := df["content_types"]; ok {
					if len(v.(*schema.Set).List()) > 0 {
						var cl []string
						for _, c := range v.(*schema.Set).List() {
							cl = append(cl, c.(string))
						}
						opts.ContentTypes = strings.Join(cl, " ")
					}
				}

				if v, ok := df["extensions"]; ok {
					if len(v.(*schema.Set).List()) > 0 {
						var el []string
						for _, e := range v.(*schema.Set).List() {
							el = append(el, e.(string))
						}
						opts.Extensions = strings.Join(el, " ")
					}
				}

				log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
				_, err := conn.CreateGzip(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in s3logging
		if d.HasChange("s3logging") {
			os, ns := d.GetChange("s3logging")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeS3Logging := oss.Difference(nss).List()
			addS3Logging := nss.Difference(oss).List()

			// DELETE old S3 Log configurations
			for _, sRaw := range removeS3Logging {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.DeleteS3Input{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly S3 Logging removal opts: %#v", opts)
				err := conn.DeleteS3(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated S3 Logging
			for _, sRaw := range addS3Logging {
				sf := sRaw.(map[string]interface{})

				// Fastly API will not error if these are omitted, so we throw an error
				// if any of these are empty
				for _, sk := range []string{"s3_access_key", "s3_secret_key"} {
					if sf[sk].(string) == "" {
						return fmt.Errorf("[ERR] No %s found for S3 Log stream setup for Service (%s)", sk, d.Id())
					}
				}

				opts := gofastly.CreateS3Input{
					Service:                      d.Id(),
					Version:                      latestVersion,
					Name:                         sf["name"].(string),
					BucketName:                   sf["bucket_name"].(string),
					AccessKey:                    sf["s3_access_key"].(string),
					SecretKey:                    sf["s3_secret_key"].(string),
					Period:                       uint(sf["period"].(int)),
					GzipLevel:                    uint(sf["gzip_level"].(int)),
					Domain:                       sf["domain"].(string),
					Path:                         sf["path"].(string),
					Format:                       sf["format"].(string),
					FormatVersion:                uint(sf["format_version"].(int)),
					TimestampFormat:              sf["timestamp_format"].(string),
					ResponseCondition:            sf["response_condition"].(string),
					MessageType:                  sf["message_type"].(string),
					Placement:                    sf["placement"].(string),
					ServerSideEncryptionKMSKeyID: sf["server_side_encryption_kms_key_id"].(string),
				}

				redundancy := strings.ToLower(sf["redundancy"].(string))
				switch redundancy {
				case "standard":
					opts.Redundancy = gofastly.S3RedundancyStandard
				case "reduced_redundancy":
					opts.Redundancy = gofastly.S3RedundancyReduced
				}

				encryption := sf["server_side_encryption"].(string)
				switch encryption {
				case string(gofastly.S3ServerSideEncryptionAES):
					opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionAES
				case string(gofastly.S3ServerSideEncryptionKMS):
					opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionKMS
				}

				log.Printf("[DEBUG] Create S3 Logging Opts: %#v", opts)
				_, err := conn.CreateS3(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Papertrail
		if d.HasChange("papertrail") {
			os, ns := d.GetChange("papertrail")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removePapertrail := oss.Difference(nss).List()
			addPapertrail := nss.Difference(oss).List()

			// DELETE old papertrail configurations
			for _, pRaw := range removePapertrail {
				pf := pRaw.(map[string]interface{})
				opts := gofastly.DeletePapertrailInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    pf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Papertrail removal opts: %#v", opts)
				err := conn.DeletePapertrail(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Papertrail
			for _, pRaw := range addPapertrail {
				pf := pRaw.(map[string]interface{})

				opts := gofastly.CreatePapertrailInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              pf["name"].(string),
					Address:           pf["address"].(string),
					Port:              uint(pf["port"].(int)),
					Format:            pf["format"].(string),
					ResponseCondition: pf["response_condition"].(string),
					Placement:         pf["placement"].(string),
				}

				log.Printf("[DEBUG] Create Papertrail Opts: %#v", opts)
				_, err := conn.CreatePapertrail(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Sumologic
		if d.HasChange("sumologic") {
			os, ns := d.GetChange("sumologic")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeSumologic := oss.Difference(nss).List()
			addSumologic := nss.Difference(oss).List()

			// DELETE old sumologic configurations
			for _, pRaw := range removeSumologic {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.DeleteSumologicInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Sumologic removal opts: %#v", opts)
				err := conn.DeleteSumologic(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Sumologic
			for _, pRaw := range addSumologic {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.CreateSumologicInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              sf["name"].(string),
					URL:               sf["url"].(string),
					Format:            sf["format"].(string),
					FormatVersion:     sf["format_version"].(int),
					ResponseCondition: sf["response_condition"].(string),
					MessageType:       sf["message_type"].(string),
					Placement:         sf["placement"].(string),
				}

				log.Printf("[DEBUG] Create Sumologic Opts: %#v", opts)
				_, err := conn.CreateSumologic(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in gcslogging
		if d.HasChange("gcslogging") {
			os, ns := d.GetChange("gcslogging")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeGcslogging := oss.Difference(nss).List()
			addGcslogging := nss.Difference(oss).List()

			// DELETE old gcslogging configurations
			for _, pRaw := range removeGcslogging {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.DeleteGCSInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly gcslogging removal opts: %#v", opts)
				err := conn.DeleteGCS(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated gcslogging
			for _, pRaw := range addGcslogging {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.CreateGCSInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              sf["name"].(string),
					User:              sf["email"].(string),
					Bucket:            sf["bucket_name"].(string),
					SecretKey:         sf["secret_key"].(string),
					Format:            sf["format"].(string),
					Path:              sf["path"].(string),
					Period:            uint(sf["period"].(int)),
					GzipLevel:         uint8(sf["gzip_level"].(int)),
					TimestampFormat:   sf["timestamp_format"].(string),
					MessageType:       sf["message_type"].(string),
					ResponseCondition: sf["response_condition"].(string),
					Placement:         sf["placement"].(string),
				}

				log.Printf("[DEBUG] Create GCS Opts: %#v", opts)
				_, err := conn.CreateGCS(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in bigquerylogging
		if d.HasChange("bigquerylogging") {
			os, ns := d.GetChange("bigquerylogging")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeBigquerylogging := oss.Difference(nss).List()
			addBigquerylogging := nss.Difference(oss).List()

			// DELETE old bigquerylogging configurations
			for _, pRaw := range removeBigquerylogging {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.DeleteBigQueryInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly bigquerylogging removal opts: %#v", opts)
				err := conn.DeleteBigQuery(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated bigquerylogging
			for _, pRaw := range addBigquerylogging {
				sf := pRaw.(map[string]interface{})
				opts := gofastly.CreateBigQueryInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              sf["name"].(string),
					ProjectID:         sf["project_id"].(string),
					Dataset:           sf["dataset"].(string),
					Table:             sf["table"].(string),
					User:              sf["email"].(string),
					SecretKey:         sf["secret_key"].(string),
					ResponseCondition: sf["response_condition"].(string),
					Template:          sf["template"].(string),
					Placement:         sf["placement"].(string),
				}

				if sf["format"].(string) != "" {
					opts.Format = sf["format"].(string)
				}

				log.Printf("[DEBUG] Create bigquerylogging opts: %#v", opts)
				_, err := conn.CreateBigQuery(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Syslog
		if d.HasChange("syslog") {
			os, ns := d.GetChange("syslog")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeSyslog := oss.Difference(nss).List()
			addSyslog := nss.Difference(oss).List()

			// DELETE old syslog configurations
			for _, pRaw := range removeSyslog {
				slf := pRaw.(map[string]interface{})
				opts := gofastly.DeleteSyslogInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    slf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Syslog removal opts: %#v", opts)
				err := conn.DeleteSyslog(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Syslog
			for _, pRaw := range addSyslog {
				slf := pRaw.(map[string]interface{})

				opts := gofastly.CreateSyslogInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              slf["name"].(string),
					Address:           slf["address"].(string),
					Port:              uint(slf["port"].(int)),
					Format:            slf["format"].(string),
					FormatVersion:     uint(slf["format_version"].(int)),
					Token:             slf["token"].(string),
					UseTLS:            gofastly.CBool(slf["use_tls"].(bool)),
					TLSHostname:       slf["tls_hostname"].(string),
					TLSCACert:         slf["tls_ca_cert"].(string),
					TLSClientCert:     slf["tls_client_cert"].(string),
					TLSClientKey:      slf["tls_client_key"].(string),
					ResponseCondition: slf["response_condition"].(string),
					MessageType:       slf["message_type"].(string),
					Placement:         slf["placement"].(string),
				}

				log.Printf("[DEBUG] Create Syslog Opts: %#v", opts)
				_, err := conn.CreateSyslog(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Logentries
		if d.HasChange("logentries") {
			os, ns := d.GetChange("logentries")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeLogentries := oss.Difference(nss).List()
			addLogentries := nss.Difference(oss).List()

			// DELETE old logentries configurations
			for _, pRaw := range removeLogentries {
				slf := pRaw.(map[string]interface{})
				opts := gofastly.DeleteLogentriesInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    slf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Logentries removal opts: %#v", opts)
				err := conn.DeleteLogentries(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Logentries
			for _, pRaw := range addLogentries {
				slf := pRaw.(map[string]interface{})

				opts := gofastly.CreateLogentriesInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              slf["name"].(string),
					Port:              uint(slf["port"].(int)),
					UseTLS:            gofastly.CBool(slf["use_tls"].(bool)),
					Token:             slf["token"].(string),
					Format:            slf["format"].(string),
					FormatVersion:     uint(slf["format_version"].(int)),
					ResponseCondition: slf["response_condition"].(string),
					Placement:         slf["placement"].(string),
				}

				log.Printf("[DEBUG] Create Logentries Opts: %#v", opts)
				_, err := conn.CreateLogentries(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Splunk logging configurations
		if d.HasChange("splunk") {
			os, ns := d.GetChange("splunk")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)

			remove := oss.Difference(nss).List()
			add := nss.Difference(oss).List()

			// DELETE old Splunk logging configurations
			for _, sRaw := range remove {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.DeleteSplunkInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Splunk removal opts: %#v", opts)
				err := conn.DeleteSplunk(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Splunk configurations
			for _, sRaw := range add {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.CreateSplunkInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              sf["name"].(string),
					URL:               sf["url"].(string),
					Format:            sf["format"].(string),
					FormatVersion:     uint(sf["format_version"].(int)),
					ResponseCondition: sf["response_condition"].(string),
					Placement:         sf["placement"].(string),
					Token:             sf["token"].(string),
					TLSHostname:       sf["tls_hostname"].(string),
					TLSCACert:         sf["tls_ca_cert"].(string),
				}

				log.Printf("[DEBUG] Splunk create opts: %#v", opts)
				_, err := conn.CreateSplunk(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in Blob Storage logging configurations
		if d.HasChange("blobstoragelogging") {
			obsl, nbsl := d.GetChange("blobstoragelogging")
			if obsl == nil {
				obsl = new(schema.Set)
			}
			if nbsl == nil {
				nbsl = new(schema.Set)
			}

			obsls := obsl.(*schema.Set)
			nbsls := nbsl.(*schema.Set)

			remove := obsls.Difference(nbsls).List()
			add := nbsls.Difference(obsls).List()

			// DELETE old Blob Storage logging configurations
			for _, bslRaw := range remove {
				bslf := bslRaw.(map[string]interface{})
				opts := gofastly.DeleteBlobStorageInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    bslf["name"].(string),
				}

				log.Printf("[DEBUG] Blob Storage logging removal opts: %#v", opts)
				err := conn.DeleteBlobStorage(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Blob Storage logging configurations
			for _, bslRaw := range add {
				bslf := bslRaw.(map[string]interface{})
				opts := gofastly.CreateBlobStorageInput{
					Service:           d.Id(),
					Version:           latestVersion,
					Name:              bslf["name"].(string),
					Path:              bslf["path"].(string),
					AccountName:       bslf["account_name"].(string),
					Container:         bslf["container"].(string),
					SASToken:          bslf["sas_token"].(string),
					Period:            uint(bslf["period"].(int)),
					TimestampFormat:   bslf["timestamp_format"].(string),
					GzipLevel:         uint(bslf["gzip_level"].(int)),
					PublicKey:         bslf["public_key"].(string),
					Format:            bslf["format"].(string),
					FormatVersion:     uint(bslf["format_version"].(int)),
					MessageType:       bslf["message_type"].(string),
					Placement:         bslf["placement"].(string),
					ResponseCondition: bslf["response_condition"].(string),
				}

				log.Printf("[DEBUG] Blob Storage logging create opts: %#v", opts)
				_, err := conn.CreateBlobStorage(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find differences in HTTPS logging configuration
		if d.HasChange("httpslogging") {
			if err := processHTTPS(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Response Object
		if d.HasChange("response_object") {
			or, nr := d.GetChange("response_object")
			if or == nil {
				or = new(schema.Set)
			}
			if nr == nil {
				nr = new(schema.Set)
			}

			ors := or.(*schema.Set)
			nrs := nr.(*schema.Set)
			removeResponseObject := ors.Difference(nrs).List()
			addResponseObject := nrs.Difference(ors).List()

			// DELETE old response object configurations
			for _, rRaw := range removeResponseObject {
				rf := rRaw.(map[string]interface{})
				opts := gofastly.DeleteResponseObjectInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    rf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Response Object removal opts: %#v", opts)
				err := conn.DeleteResponseObject(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Response Object
			for _, rRaw := range addResponseObject {
				rf := rRaw.(map[string]interface{})

				opts := gofastly.CreateResponseObjectInput{
					Service:          d.Id(),
					Version:          latestVersion,
					Name:             rf["name"].(string),
					Status:           uint(rf["status"].(int)),
					Response:         rf["response"].(string),
					Content:          rf["content"].(string),
					ContentType:      rf["content_type"].(string),
					RequestCondition: rf["request_condition"].(string),
					CacheCondition:   rf["cache_condition"].(string),
				}

				log.Printf("[DEBUG] Create Response Object Opts: %#v", opts)
				_, err := conn.CreateResponseObject(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in request settings
		if d.HasChange("request_setting") {
			os, ns := d.GetChange("request_setting")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			ors := os.(*schema.Set)
			nrs := ns.(*schema.Set)
			removeRequestSettings := ors.Difference(nrs).List()
			addRequestSettings := nrs.Difference(ors).List()

			// DELETE old Request Settings configurations
			for _, sRaw := range removeRequestSettings {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.DeleteRequestSettingInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Request Setting removal opts: %#v", opts)
				err := conn.DeleteRequestSetting(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new/updated Request Setting
			for _, sRaw := range addRequestSettings {
				opts, err := buildRequestSetting(sRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Requset Setting: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
				_, err = conn.CreateRequestSetting(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in VCLs
		if d.HasChange("vcl") {
			// Note: as above with Gzip and S3 logging, we don't utilize the PUT
			// endpoint to update a VCL, we simply destroy it and create a new one.
			oldVCLVal, newVCLVal := d.GetChange("vcl")
			if oldVCLVal == nil {
				oldVCLVal = new(schema.Set)
			}
			if newVCLVal == nil {
				newVCLVal = new(schema.Set)
			}

			oldVCLSet := oldVCLVal.(*schema.Set)
			newVCLSet := newVCLVal.(*schema.Set)

			remove := oldVCLSet.Difference(newVCLSet).List()
			add := newVCLSet.Difference(oldVCLSet).List()

			// Delete removed VCL configurations
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteVCLInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Removal opts: %#v", opts)
				err := conn.DeleteVCL(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}
			// POST new VCL configurations
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateVCLInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
					Content: df["content"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Addition opts: %#v", opts)
				_, err := conn.CreateVCL(&opts)
				if err != nil {
					return err
				}

				// if this new VCL is the main
				if df["main"].(bool) {
					opts := gofastly.ActivateVCLInput{
						Service: d.Id(),
						Version: latestVersion,
						Name:    df["name"].(string),
					}
					log.Printf("[DEBUG] Fastly VCL activation opts: %#v", opts)
					_, err := conn.ActivateVCL(&opts)
					if err != nil {
						return err
					}

				}
			}
		}

		// Find differences in VCL snippets
		if d.HasChange("snippet") {
			// Note: as above with Gzip and S3 logging, we don't utilize the PUT
			// endpoint to update a VCL snippet, we simply destroy it and create a new one.
			oldSnippetVal, newSnippetVal := d.GetChange("snippet")
			if oldSnippetVal == nil {
				oldSnippetVal = new(schema.Set)
			}
			if newSnippetVal == nil {
				newSnippetVal = new(schema.Set)
			}

			oldSnippetSet := oldSnippetVal.(*schema.Set)
			newSnippetSet := newSnippetVal.(*schema.Set)

			remove := oldSnippetSet.Difference(newSnippetSet).List()
			add := newSnippetSet.Difference(oldSnippetSet).List()

			// Delete removed VCL Snippet configurations
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteSnippetInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Snippet Removal opts: %#v", opts)
				err := conn.DeleteSnippet(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new VCL Snippet configurations
			for _, dRaw := range add {
				opts, err := buildSnippet(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building VCL Snippet: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly VCL Snippet Addition opts: %#v", opts)
				_, err = conn.CreateSnippet(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in VCL dynamic snippets
		if d.HasChange("dynamicsnippet") {
			// Note: as above with Gzip and S3 logging, we don't utilize the PUT
			// endpoint to update a VCL dynamic snippet, we simply destroy it and create a new one.
			oldDynamicSnippetVal, newDynamicSnippetVal := d.GetChange("dynamicsnippet")
			if oldDynamicSnippetVal == nil {
				oldDynamicSnippetVal = new(schema.Set)
			}
			if newDynamicSnippetVal == nil {
				newDynamicSnippetVal = new(schema.Set)
			}

			oldDynamicSnippetSet := oldDynamicSnippetVal.(*schema.Set)
			newDynamicSnippetSet := newDynamicSnippetVal.(*schema.Set)

			remove := oldDynamicSnippetSet.Difference(newDynamicSnippetSet).List()
			add := newDynamicSnippetSet.Difference(oldDynamicSnippetSet).List()

			// Delete removed VCL Snippet configurations
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteSnippetInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Removal opts: %#v", opts)
				err := conn.DeleteSnippet(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new VCL Snippet configurations
			for _, dRaw := range add {
				opts, err := buildDynamicSnippet(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building VCL Dynamic Snippet: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Addition opts: %#v", opts)
				_, err = conn.CreateSnippet(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in Cache Settings
		if d.HasChange("cache_setting") {
			oc, nc := d.GetChange("cache_setting")
			if oc == nil {
				oc = new(schema.Set)
			}
			if nc == nil {
				nc = new(schema.Set)
			}

			ocs := oc.(*schema.Set)
			ncs := nc.(*schema.Set)

			remove := ocs.Difference(ncs).List()
			add := ncs.Difference(ocs).List()

			// Delete removed Cache Settings
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteCacheSettingInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Cache Settings removal opts: %#v", opts)
				err := conn.DeleteCacheSetting(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new Cache Settings
			for _, dRaw := range add {
				opts, err := buildCacheSetting(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Cache Setting: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly Cache Settings Addition opts: %#v", opts)
				_, err = conn.CreateCacheSetting(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in ACLs
		if d.HasChange("acl") {
			if err := processACL(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in dictionary
		if d.HasChange("dictionary") {

			oldDictVal, newDictVal := d.GetChange("dictionary")

			if oldDictVal == nil {
				oldDictVal = new(schema.Set)
			}
			if newDictVal == nil {
				newDictVal = new(schema.Set)
			}

			oldDictSet := oldDictVal.(*schema.Set)
			newDictSet := newDictVal.(*schema.Set)

			remove := oldDictSet.Difference(newDictSet).List()
			add := newDictSet.Difference(oldDictSet).List()

			// Delete removed dictionary configurations
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteDictionaryInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Dictionary Removal opts: %#v", opts)
				err := conn.DeleteDictionary(&opts)
				if errRes, ok := err.(*gofastly.HTTPError); ok {
					if errRes.StatusCode != 404 {
						return err
					}
				} else if err != nil {
					return err
				}
			}

			// POST new dictionary configurations
			for _, dRaw := range add {
				opts, err := buildDictionary(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Dicitionary: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly Dictionary Addition opts: %#v", opts)
				_, err = conn.CreateDictionary(opts)
				if err != nil {
					return err
				}
			}
		}

		// validate version
		log.Printf("[DEBUG] Validating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
		valid, msg, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error checking validation: %s", err)
		}

		if !valid {
			return fmt.Errorf("[ERR] Invalid configuration for Fastly Service (%s): %s", d.Id(), msg)
		}

		shouldActivate := d.Get("activate").(bool)
		if shouldActivate {
			log.Printf("[DEBUG] Activating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return fmt.Errorf("[ERR] Error activating version (%d): %s", latestVersion, err)
			}

			// Only if the version is valid and activated do we set the active_version.
			// This prevents us from getting stuck in cloning an invalid version
			d.Set("active_version", latestVersion)
		} else {
			log.Printf("[INFO] Skipping activation of Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			log.Print("[INFO] The Terraform definition is explicitly specified to not activate the changes on Fastly")
			log.Printf("[INFO] Version (%v) has been pushed and validated", latestVersion)
			log.Printf("[INFO] Visit https://manage.fastly.com/configure/services/%s/versions/%v and activate it manually", d.Id(), latestVersion)
		}
	}

	return resourceServiceV1Read(d, meta)
}

func resourceServiceV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Find the Service. Discard the service because we need the ServiceDetails,
	// not just a Service record
	_, err := findService(d.Id(), meta)
	if err != nil {
		switch err {
		case fastlyNoServiceFoundErr:
			log.Printf("[WARN] %s for ID (%s)", err, d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("name", s.Name)
	d.Set("comment", s.Comment)
	d.Set("version_comment", s.Version.Comment)
	d.Set("active_version", s.ActiveVersion.Number)

	// If CreateService succeeds, but initial updates to the Service fail, we'll
	// have an empty ActiveService version (no version is active, so we can't
	// query for information on it)
	if s.ActiveVersion.Number != 0 {
		settingsOpts := gofastly.GetSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		}
		if settings, err := conn.GetSettings(&settingsOpts); err == nil {
			d.Set("default_host", settings.DefaultHost)
			d.Set("default_ttl", settings.DefaultTTL)
		} else {
			return fmt.Errorf("[ERR] Error looking up Version settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		// TODO: update go-fastly to support an ActiveVersion struct, which contains
		// domain and backend info in the response. Here we do 2 additional queries
		// to find out that info
		log.Printf("[DEBUG] Refreshing Domains for (%s)", d.Id())
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		// Refresh Domains
		dl := flattenDomains(domainList)

		if err := d.Set("domain", dl); err != nil {
			log.Printf("[WARN] Error setting Domains for (%s): %s", d.Id(), err)
		}

		// Refresh Backends
		log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		bl := flattenBackends(backendList)

		if err := d.Set("backend", bl); err != nil {
			log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
		}

		// refresh directors
		log.Printf("[DEBUG] Refreshing Directors for (%s)", d.Id())
		directorList, err := conn.ListDirectors(&gofastly.ListDirectorsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Directors for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		log.Printf("[DEBUG] Refreshing Director Backends for (%s)", d.Id())
		var directorBackendList []*gofastly.DirectorBackend

		for _, director := range directorList {
			for _, backend := range backendList {
				directorBackendGet, err := conn.GetDirectorBackend(&gofastly.GetDirectorBackendInput{
					Service:  d.Id(),
					Version:  s.ActiveVersion.Number,
					Director: director.Name,
					Backend:  backend.Name,
				})
				if err == nil {
					directorBackendList = append(directorBackendList, directorBackendGet)
				}
			}
		}

		dirl := flattenDirectors(directorList, directorBackendList)

		if err := d.Set("director", dirl); err != nil {
			log.Printf("[WARN] Error setting Directors for (%s): %s", d.Id(), err)
		}

		// refresh headers
		log.Printf("[DEBUG] Refreshing Headers for (%s)", d.Id())
		headerList, err := conn.ListHeaders(&gofastly.ListHeadersInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Headers for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		hl := flattenHeaders(headerList)

		if err := d.Set("header", hl); err != nil {
			log.Printf("[WARN] Error setting Headers for (%s): %s", d.Id(), err)
		}

		// refresh gzips
		log.Printf("[DEBUG] Refreshing Gzips for (%s)", d.Id())
		gzipsList, err := conn.ListGzips(&gofastly.ListGzipsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Gzips for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		gl := flattenGzips(gzipsList)

		if err := d.Set("gzip", gl); err != nil {
			log.Printf("[WARN] Error setting Gzips for (%s): %s", d.Id(), err)
		}

		// refresh Healthcheck
		log.Printf("[DEBUG] Refreshing Healthcheck for (%s)", d.Id())
		healthcheckList, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Healthcheck for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		hcl := flattenHealthchecks(healthcheckList)

		if err := d.Set("healthcheck", hcl); err != nil {
			log.Printf("[WARN] Error setting Healthcheck for (%s): %s", d.Id(), err)
		}

		// refresh S3 Logging
		log.Printf("[DEBUG] Refreshing S3 Logging for (%s)", d.Id())
		s3List, err := conn.ListS3s(&gofastly.ListS3sInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up S3 Logging for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		sl := flattenS3s(s3List)

		if err := d.Set("s3logging", sl); err != nil {
			log.Printf("[WARN] Error setting S3 Logging for (%s): %s", d.Id(), err)
		}

		// refresh Papertrail Logging
		log.Printf("[DEBUG] Refreshing Papertrail for (%s)", d.Id())
		papertrailList, err := conn.ListPapertrails(&gofastly.ListPapertrailsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Papertrail for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		pl := flattenPapertrails(papertrailList)

		if err := d.Set("papertrail", pl); err != nil {
			log.Printf("[WARN] Error setting Papertrail for (%s): %s", d.Id(), err)
		}

		// refresh Sumologic Logging
		log.Printf("[DEBUG] Refreshing Sumologic for (%s)", d.Id())
		sumologicList, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Sumologic for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		sul := flattenSumologics(sumologicList)
		if err := d.Set("sumologic", sul); err != nil {
			log.Printf("[WARN] Error setting Sumologic for (%s): %s", d.Id(), err)
		}

		// refresh GCS Logging
		log.Printf("[DEBUG] Refreshing GCS for (%s)", d.Id())
		GCSList, err := conn.ListGCSs(&gofastly.ListGCSsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up GCS for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		gcsl := flattenGCS(GCSList)
		if err := d.Set("gcslogging", gcsl); err != nil {
			log.Printf("[WARN] Error setting gcs for (%s): %s", d.Id(), err)
		}

		// refresh BigQuery Logging
		log.Printf("[DEBUG] Refreshing BigQuery for (%s)", d.Id())
		BQList, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up BigQuery logging for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		bql := flattenBigQuery(BQList)
		if err := d.Set("bigquerylogging", bql); err != nil {
			log.Printf("[WARN] Error setting bigquerylogging for (%s): %s", d.Id(), err)
		}

		// refresh Syslog Logging
		log.Printf("[DEBUG] Refreshing Syslog for (%s)", d.Id())
		syslogList, err := conn.ListSyslogs(&gofastly.ListSyslogsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Syslog for (%s), version (%d): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		sll := flattenSyslogs(syslogList)

		if err := d.Set("syslog", sll); err != nil {
			log.Printf("[WARN] Error setting Syslog for (%s): %s", d.Id(), err)
		}

		// refresh Logentries Logging
		log.Printf("[DEBUG] Refreshing Logentries for (%s)", d.Id())
		logentriesList, err := conn.ListLogentries(&gofastly.ListLogentriesInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Logentries for (%s), version (%d): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		lel := flattenLogentries(logentriesList)

		if err := d.Set("logentries", lel); err != nil {
			log.Printf("[WARN] Error setting Logentries for (%s): %s", d.Id(), err)
		}

		// refresh Splunk Logging
		log.Printf("[DEBUG] Refreshing Splunks for (%s)", d.Id())
		splunkList, err := conn.ListSplunks(&gofastly.ListSplunksInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Splunks for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		spl := flattenSplunks(splunkList)

		if err := d.Set("splunk", spl); err != nil {
			log.Printf("[WARN] Error setting Splunks for (%s): %s", d.Id(), err)
		}

		// refresh Blob Storage Logging
		log.Printf("[DEBUG] Refreshing Blob Storages for (%s)", d.Id())
		blobStorageList, err := conn.ListBlobStorages(&gofastly.ListBlobStoragesInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Blob Storages for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		bsl := flattenBlobStorages(blobStorageList)

		if err := d.Set("blobstoragelogging", bsl); err != nil {
			log.Printf("[WARN] Error setting Blob Storages for (%s): %s", d.Id(), err)
		}

		// Refresh HTTPS
		if err := readHTTPS(conn, d, s); err != nil {
			return err
		}

		// refresh Response Objects
		log.Printf("[DEBUG] Refreshing Response Object for (%s)", d.Id())
		responseObjectList, err := conn.ListResponseObjects(&gofastly.ListResponseObjectsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Response Object for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		rol := flattenResponseObjects(responseObjectList)

		if err := d.Set("response_object", rol); err != nil {
			log.Printf("[WARN] Error setting Response Object for (%s): %s", d.Id(), err)
		}

		// refresh Conditions
		log.Printf("[DEBUG] Refreshing Conditions for (%s)", d.Id())
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Conditions for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		cl := flattenConditions(conditionList)

		if err := d.Set("condition", cl); err != nil {
			log.Printf("[WARN] Error setting Conditions for (%s): %s", d.Id(), err)
		}

		// refresh Request Settings
		log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
		rsList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Request Settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		rl := flattenRequestSettings(rsList)

		if err := d.Set("request_setting", rl); err != nil {
			log.Printf("[WARN] Error setting Request Settings for (%s): %s", d.Id(), err)
		}

		// refresh VCLs
		log.Printf("[DEBUG] Refreshing VCLs for (%s)", d.Id())
		vclList, err := conn.ListVCLs(&gofastly.ListVCLsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up VCLs for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		vl := flattenVCLs(vclList)

		if err := d.Set("vcl", vl); err != nil {
			log.Printf("[WARN] Error setting VCLs for (%s): %s", d.Id(), err)
		}

		// refresh ACLs
		if err := readACL(conn, d, s); err != nil {
			return err
		}

		// refresh VCL Snippets
		log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
		snippetList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		vsl := flattenSnippets(snippetList)

		if err := d.Set("snippet", vsl); err != nil {
			log.Printf("[WARN] Error setting VCL Snippets for (%s): %s", d.Id(), err)
		}

		dynamicSnippets := flattenDynamicSnippets(snippetList)

		if err := d.Set("dynamicsnippet", dynamicSnippets); err != nil {
			log.Printf("[WARN] Error setting VCL Dynamic Snippets for (%s): %s", d.Id(), err)
		}

		// refresh Cache Settings
		log.Printf("[DEBUG] Refreshing Cache Settings for (%s)", d.Id())
		cslList, err := conn.ListCacheSettings(&gofastly.ListCacheSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Cache Settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		csl := flattenCacheSettings(cslList)

		if err := d.Set("cache_setting", csl); err != nil {
			log.Printf("[WARN] Error setting Cache Settings for (%s): %s", d.Id(), err)
		}

		// refresh Dictionaries
		log.Printf("[DEBUG] Refreshing Dictionaries for (%s)", d.Id())
		dictList, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionaries for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		dict := flattenDictionaries(dictList)

		if err := d.Set("dictionary", dict); err != nil {
			log.Printf("[WARN] Error setting Dictionary for (%s): %s", d.Id(), err)
		}

	} else {
		log.Printf("[DEBUG] Active Version for Service (%s) is empty, no state to refresh", d.Id())
	}

	return nil
}

func resourceServiceV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Fastly will fail to delete any service with an Active Version.
	// If `force_destroy` is given, we deactivate the active version and then send
	// the DELETE call
	if d.Get("force_destroy").(bool) {
		s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: d.Id(),
		})

		if err != nil {
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
				Service: d.Id(),
				Version: s.ActiveVersion.Number,
			})
			if err != nil {
				return err
			}
		}
	}

	err := conn.DeleteService(&gofastly.DeleteServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	_, err = findService(d.Id(), meta)
	if err != nil {
		switch err {
		// we expect no records to be found here
		case fastlyNoServiceFoundErr:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// findService above returned something and nil error, but shouldn't have
	return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", d.Id())

}

// findService finds a Fastly Service via the ListServices endpoint, returning
// the Service if found.
//
// Fastly API does not include any "deleted_at" type parameter to indicate
// that a Service has been deleted. GET requests to a deleted Service will
// return 200 OK and have the full output of the Service for an unknown time
// (days, in my testing). In order to determine if a Service is deleted, we
// need to hit /service and loop the returned Services, searching for the one
// in question. This endpoint only returns active or "alive" services. If the
// Service is not included, then it's "gone"
//
// Returns a fastlyNoServiceFoundErr error if the Service is not found in the
// ListServices response.
func findService(id string, meta interface{}) (*gofastly.Service, error) {
	conn := meta.(*FastlyClient).conn

	l, err := conn.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("[WARN] Error listing services (%s): %s", id, err)
	}

	for _, s := range l {
		if s.ID == id {
			log.Printf("[DEBUG] Found Service (%s)", id)
			return s, nil
		}
	}

	return nil, fastlyNoServiceFoundErr
}

func flattenDomains(list []*gofastly.Domain) []map[string]interface{} {
	dl := make([]map[string]interface{}, 0, len(list))

	for _, d := range list {
		dl = append(dl, map[string]interface{}{
			"name":    d.Name,
			"comment": d.Comment,
		})
	}

	return dl
}

func flattenBackends(backendList []*gofastly.Backend) []map[string]interface{} {
	var bl []map[string]interface{}
	for _, b := range backendList {
		// Convert Backend to a map for saving to state.
		nb := map[string]interface{}{
			"name":                  b.Name,
			"address":               b.Address,
			"auto_loadbalance":      b.AutoLoadbalance,
			"between_bytes_timeout": int(b.BetweenBytesTimeout),
			"connect_timeout":       int(b.ConnectTimeout),
			"error_threshold":       int(b.ErrorThreshold),
			"first_byte_timeout":    int(b.FirstByteTimeout),
			"max_conn":              int(b.MaxConn),
			"port":                  int(b.Port),
			"override_host":         b.OverrideHost,
			"shield":                b.Shield,
			"ssl_check_cert":        b.SSLCheckCert,
			"ssl_hostname":          b.SSLHostname,
			"ssl_ca_cert":           b.SSLCACert,
			"ssl_client_key":        b.SSLClientKey,
			"ssl_client_cert":       b.SSLClientCert,
			"max_tls_version":       b.MaxTLSVersion,
			"min_tls_version":       b.MinTLSVersion,
			"ssl_ciphers":           strings.Join(b.SSLCiphers, ","),
			"use_ssl":               b.UseSSL,
			"ssl_cert_hostname":     b.SSLCertHostname,
			"ssl_sni_hostname":      b.SSLSNIHostname,
			"weight":                int(b.Weight),
			"request_condition":     b.RequestCondition,
			"healthcheck":           b.HealthCheck,
		}

		bl = append(bl, nb)
	}
	return bl
}

func flattenDirectors(directorList []*gofastly.Director, directorBackendList []*gofastly.DirectorBackend) []map[string]interface{} {
	var dl []map[string]interface{}
	for _, d := range directorList {
		// Convert Director to a map for saving to state.
		nd := map[string]interface{}{
			"name":     d.Name,
			"comment":  d.Comment,
			"shield":   d.Shield,
			"type":     d.Type,
			"quorum":   int(d.Quorum),
			"capacity": int(d.Capacity),
			"retries":  int(d.Retries),
		}

		var b []interface{}
		for _, db := range directorBackendList {
			if d.Name == db.Director {
				b = append(b, db.Backend)
			}
		}
		if len(b) > 0 {
			nd["backends"] = schema.NewSet(schema.HashString, b)
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nd {
			if v == "" {
				delete(nd, k)
			}
		}

		dl = append(dl, nd)
	}
	return dl
}

func flattenHeaders(headerList []*gofastly.Header) []map[string]interface{} {
	var hl []map[string]interface{}
	for _, h := range headerList {
		// Convert Header to a map for saving to state.
		nh := map[string]interface{}{
			"name":               h.Name,
			"action":             h.Action,
			"ignore_if_set":      h.IgnoreIfSet,
			"type":               h.Type,
			"destination":        h.Destination,
			"source":             h.Source,
			"regex":              h.Regex,
			"substitution":       h.Substitution,
			"priority":           int(h.Priority),
			"request_condition":  h.RequestCondition,
			"cache_condition":    h.CacheCondition,
			"response_condition": h.ResponseCondition,
		}

		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}
	return hl
}

func buildHeader(headerMap interface{}) (*gofastly.CreateHeaderInput, error) {
	df := headerMap.(map[string]interface{})
	opts := gofastly.CreateHeaderInput{
		Name:              df["name"].(string),
		IgnoreIfSet:       gofastly.CBool(df["ignore_if_set"].(bool)),
		Destination:       df["destination"].(string),
		Priority:          uint(df["priority"].(int)),
		Source:            df["source"].(string),
		Regex:             df["regex"].(string),
		Substitution:      df["substitution"].(string),
		RequestCondition:  df["request_condition"].(string),
		CacheCondition:    df["cache_condition"].(string),
		ResponseCondition: df["response_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "set":
		opts.Action = gofastly.HeaderActionSet
	case "append":
		opts.Action = gofastly.HeaderActionAppend
	case "delete":
		opts.Action = gofastly.HeaderActionDelete
	case "regex":
		opts.Action = gofastly.HeaderActionRegex
	case "regex_repeat":
		opts.Action = gofastly.HeaderActionRegexRepeat
	}

	ty := strings.ToLower(df["type"].(string))
	switch ty {
	case "request":
		opts.Type = gofastly.HeaderTypeRequest
	case "fetch":
		opts.Type = gofastly.HeaderTypeFetch
	case "cache":
		opts.Type = gofastly.HeaderTypeCache
	case "response":
		opts.Type = gofastly.HeaderTypeResponse
	}

	return &opts, nil
}

func buildCacheSetting(cacheMap interface{}) (*gofastly.CreateCacheSettingInput, error) {
	df := cacheMap.(map[string]interface{})
	opts := gofastly.CreateCacheSettingInput{
		Name:           df["name"].(string),
		StaleTTL:       uint(df["stale_ttl"].(int)),
		CacheCondition: df["cache_condition"].(string),
	}

	if v, ok := df["ttl"]; ok {
		opts.TTL = uint(v.(int))
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "cache":
		opts.Action = gofastly.CacheSettingActionCache
	case "pass":
		opts.Action = gofastly.CacheSettingActionPass
	case "restart":
		opts.Action = gofastly.CacheSettingActionRestart
	}

	return &opts, nil
}

func flattenGzips(gzipsList []*gofastly.Gzip) []map[string]interface{} {
	var gl []map[string]interface{}
	for _, g := range gzipsList {
		// Convert Gzip to a map for saving to state.
		ng := map[string]interface{}{
			"name":            g.Name,
			"cache_condition": g.CacheCondition,
		}

		if g.Extensions != "" {
			e := strings.Split(g.Extensions, " ")
			var et []interface{}
			for _, ev := range e {
				et = append(et, ev)
			}
			ng["extensions"] = schema.NewSet(schema.HashString, et)
		}

		if g.ContentTypes != "" {
			c := strings.Split(g.ContentTypes, " ")
			var ct []interface{}
			for _, cv := range c {
				ct = append(ct, cv)
			}
			ng["content_types"] = schema.NewSet(schema.HashString, ct)
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ng {
			if v == "" {
				delete(ng, k)
			}
		}

		gl = append(gl, ng)
	}

	return gl
}

func flattenHealthchecks(healthcheckList []*gofastly.HealthCheck) []map[string]interface{} {
	var hl []map[string]interface{}
	for _, h := range healthcheckList {
		// Convert HealthChecks to a map for saving to state.
		nh := map[string]interface{}{
			"name":              h.Name,
			"host":              h.Host,
			"path":              h.Path,
			"check_interval":    h.CheckInterval,
			"expected_response": h.ExpectedResponse,
			"http_version":      h.HTTPVersion,
			"initial":           h.Initial,
			"method":            h.Method,
			"threshold":         h.Threshold,
			"timeout":           h.Timeout,
			"window":            h.Window,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}

	return hl
}

func flattenS3s(s3List []*gofastly.S3) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, s := range s3List {
		// Convert S3s to a map for saving to state.
		ns := map[string]interface{}{
			"name":                              s.Name,
			"bucket_name":                       s.BucketName,
			"s3_access_key":                     s.AccessKey,
			"s3_secret_key":                     s.SecretKey,
			"path":                              s.Path,
			"period":                            s.Period,
			"domain":                            s.Domain,
			"gzip_level":                        s.GzipLevel,
			"format":                            s.Format,
			"format_version":                    s.FormatVersion,
			"timestamp_format":                  s.TimestampFormat,
			"redundancy":                        s.Redundancy,
			"response_condition":                s.ResponseCondition,
			"message_type":                      s.MessageType,
			"placement":                         s.Placement,
			"server_side_encryption":            s.ServerSideEncryption,
			"server_side_encryption_kms_key_id": s.ServerSideEncryptionKMSKeyID,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		sl = append(sl, ns)
	}

	return sl
}

func flattenPapertrails(papertrailList []*gofastly.Papertrail) []map[string]interface{} {
	var pl []map[string]interface{}
	for _, p := range papertrailList {
		// Convert Papertrails to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"address":            p.Address,
			"port":               p.Port,
			"format":             p.Format,
			"response_condition": p.ResponseCondition,
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		pl = append(pl, ns)
	}

	return pl
}

func flattenSumologics(sumologicList []*gofastly.Sumologic) []map[string]interface{} {
	var l []map[string]interface{}
	for _, p := range sumologicList {
		// Convert Sumologic to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"url":                p.URL,
			"format":             p.Format,
			"response_condition": p.ResponseCondition,
			"message_type":       p.MessageType,
			"format_version":     int(p.FormatVersion),
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		l = append(l, ns)
	}

	return l
}

func flattenGCS(gcsList []*gofastly.GCS) []map[string]interface{} {
	var GCSList []map[string]interface{}
	for _, currentGCS := range gcsList {
		// Convert gcs to a map for saving to state.
		GCSMapString := map[string]interface{}{
			"name":               currentGCS.Name,
			"email":              currentGCS.User,
			"bucket_name":        currentGCS.Bucket,
			"secret_key":         currentGCS.SecretKey,
			"path":               currentGCS.Path,
			"period":             int(currentGCS.Period),
			"gzip_level":         int(currentGCS.GzipLevel),
			"response_condition": currentGCS.ResponseCondition,
			"message_type":       currentGCS.MessageType,
			"format":             currentGCS.Format,
			"timestamp_format":   currentGCS.TimestampFormat,
			"placement":          currentGCS.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range GCSMapString {
			if v == "" {
				delete(GCSMapString, k)
			}
		}

		GCSList = append(GCSList, GCSMapString)
	}

	return GCSList
}

func flattenBigQuery(bqList []*gofastly.BigQuery) []map[string]interface{} {
	var BQList []map[string]interface{}
	for _, currentBQ := range bqList {
		// Convert gcs to a map for saving to state.
		BQMapString := map[string]interface{}{
			"name":               currentBQ.Name,
			"format":             currentBQ.Format,
			"email":              currentBQ.User,
			"secret_key":         currentBQ.SecretKey,
			"project_id":         currentBQ.ProjectID,
			"dataset":            currentBQ.Dataset,
			"table":              currentBQ.Table,
			"response_condition": currentBQ.ResponseCondition,
			"template":           currentBQ.Template,
			"placement":          currentBQ.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range BQMapString {
			if v == "" {
				delete(BQMapString, k)
			}
		}

		BQList = append(BQList, BQMapString)
	}

	return BQList
}

func flattenSyslogs(syslogList []*gofastly.Syslog) []map[string]interface{} {
	var pl []map[string]interface{}
	for _, p := range syslogList {
		// Convert Syslog to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"address":            p.Address,
			"port":               p.Port,
			"format":             p.Format,
			"format_version":     p.FormatVersion,
			"token":              p.Token,
			"use_tls":            p.UseTLS,
			"tls_hostname":       p.TLSHostname,
			"tls_ca_cert":        p.TLSCACert,
			"tls_client_cert":    p.TLSClientCert,
			"tls_client_key":     p.TLSClientKey,
			"response_condition": p.ResponseCondition,
			"message_type":       p.MessageType,
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		pl = append(pl, ns)
	}

	return pl
}

func flattenLogentries(logentriesList []*gofastly.Logentries) []map[string]interface{} {
	var LEList []map[string]interface{}
	for _, currentLE := range logentriesList {
		// Convert Logentries to a map for saving to state.
		LEMapString := map[string]interface{}{
			"name":               currentLE.Name,
			"port":               currentLE.Port,
			"use_tls":            currentLE.UseTLS,
			"token":              currentLE.Token,
			"format":             currentLE.Format,
			"format_version":     currentLE.FormatVersion,
			"response_condition": currentLE.ResponseCondition,
			"placement":          currentLE.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range LEMapString {
			if v == "" {
				delete(LEMapString, k)
			}
		}

		LEList = append(LEList, LEMapString)
	}

	return LEList
}

func flattenSplunks(splunkList []*gofastly.Splunk) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, s := range splunkList {
		// Convert Splunk to a map for saving to state.
		nbs := map[string]interface{}{
			"name":               s.Name,
			"url":                s.URL,
			"format":             s.Format,
			"format_version":     s.FormatVersion,
			"response_condition": s.ResponseCondition,
			"placement":          s.Placement,
			"token":              s.Token,
			"tls_hostname":       s.TLSHostname,
			"tls_ca_cert":        s.TLSCACert,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nbs {
			if v == "" {
				delete(nbs, k)
			}
		}

		sl = append(sl, nbs)
	}

	return sl
}

func flattenBlobStorages(blobStorageList []*gofastly.BlobStorage) []map[string]interface{} {
	var bsl []map[string]interface{}
	for _, bs := range blobStorageList {
		// Convert Blob Storages to a map for saving to state.
		nbs := map[string]interface{}{
			"name":               bs.Name,
			"path":               bs.Path,
			"account_name":       bs.AccountName,
			"container":          bs.Container,
			"sas_token":          bs.SASToken,
			"period":             bs.Period,
			"timestamp_format":   bs.TimestampFormat,
			"gzip_level":         bs.GzipLevel,
			"public_key":         bs.PublicKey,
			"format":             bs.Format,
			"format_version":     bs.FormatVersion,
			"message_type":       bs.MessageType,
			"placement":          bs.Placement,
			"response_condition": bs.ResponseCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nbs {
			if v == "" {
				delete(nbs, k)
			}
		}

		bsl = append(bsl, nbs)
	}

	return bsl
}

func flattenResponseObjects(responseObjectList []*gofastly.ResponseObject) []map[string]interface{} {
	var rol []map[string]interface{}
	for _, ro := range responseObjectList {
		// Convert ResponseObjects to a map for saving to state.
		nro := map[string]interface{}{
			"name":              ro.Name,
			"status":            ro.Status,
			"response":          ro.Response,
			"content":           ro.Content,
			"content_type":      ro.ContentType,
			"request_condition": ro.RequestCondition,
			"cache_condition":   ro.CacheCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nro {
			if v == "" {
				delete(nro, k)
			}
		}

		rol = append(rol, nro)
	}

	return rol
}

func flattenConditions(conditionList []*gofastly.Condition) []map[string]interface{} {
	var cl []map[string]interface{}
	for _, c := range conditionList {
		// Convert Conditions to a map for saving to state.
		nc := map[string]interface{}{
			"name":      c.Name,
			"statement": c.Statement,
			"type":      c.Type,
			"priority":  c.Priority,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nc {
			if v == "" {
				delete(nc, k)
			}
		}

		cl = append(cl, nc)
	}

	return cl
}

func flattenRequestSettings(rsList []*gofastly.RequestSetting) []map[string]interface{} {
	var rl []map[string]interface{}
	for _, r := range rsList {
		// Convert Request Settings to a map for saving to state.
		nrs := map[string]interface{}{
			"name":              r.Name,
			"max_stale_age":     r.MaxStaleAge,
			"force_miss":        r.ForceMiss,
			"force_ssl":         r.ForceSSL,
			"action":            r.Action,
			"bypass_busy_wait":  r.BypassBusyWait,
			"hash_keys":         r.HashKeys,
			"xff":               r.XForwardedFor,
			"timer_support":     r.TimerSupport,
			"geo_headers":       r.GeoHeaders,
			"default_host":      r.DefaultHost,
			"request_condition": r.RequestCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nrs {
			if v == "" {
				delete(nrs, k)
			}
		}

		rl = append(rl, nrs)
	}

	return rl
}

func buildRequestSetting(requestSettingMap interface{}) (*gofastly.CreateRequestSettingInput, error) {
	df := requestSettingMap.(map[string]interface{})
	opts := gofastly.CreateRequestSettingInput{
		Name:             df["name"].(string),
		MaxStaleAge:      uint(df["max_stale_age"].(int)),
		ForceMiss:        gofastly.CBool(df["force_miss"].(bool)),
		ForceSSL:         gofastly.CBool(df["force_ssl"].(bool)),
		BypassBusyWait:   gofastly.CBool(df["bypass_busy_wait"].(bool)),
		HashKeys:         df["hash_keys"].(string),
		TimerSupport:     gofastly.CBool(df["timer_support"].(bool)),
		GeoHeaders:       gofastly.CBool(df["geo_headers"].(bool)),
		DefaultHost:      df["default_host"].(string),
		RequestCondition: df["request_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "lookup":
		opts.Action = gofastly.RequestSettingActionLookup
	case "pass":
		opts.Action = gofastly.RequestSettingActionPass
	}

	xff := strings.ToLower(df["xff"].(string))
	switch xff {
	case "clear":
		opts.XForwardedFor = gofastly.RequestSettingXFFClear
	case "leave":
		opts.XForwardedFor = gofastly.RequestSettingXFFLeave
	case "append":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppend
	case "append_all":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppendAll
	case "overwrite":
		opts.XForwardedFor = gofastly.RequestSettingXFFOverwrite
	}

	return &opts, nil
}

func flattenCacheSettings(csList []*gofastly.CacheSetting) []map[string]interface{} {
	var csl []map[string]interface{}
	for _, cl := range csList {
		// Convert Cache Settings to a map for saving to state.
		clMap := map[string]interface{}{
			"name":            cl.Name,
			"action":          cl.Action,
			"cache_condition": cl.CacheCondition,
			"stale_ttl":       cl.StaleTTL,
			"ttl":             cl.TTL,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range clMap {
			if v == "" {
				delete(clMap, k)
			}
		}

		csl = append(csl, clMap)
	}

	return csl
}

func flattenVCLs(vclList []*gofastly.VCL) []map[string]interface{} {
	var vl []map[string]interface{}
	for _, vcl := range vclList {
		// Convert VCLs to a map for saving to state.
		vclMap := map[string]interface{}{
			"name":    vcl.Name,
			"content": vcl.Content,
			"main":    vcl.Main,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range vclMap {
			if v == "" {
				delete(vclMap, k)
			}
		}

		vl = append(vl, vclMap)
	}

	return vl
}


func buildSnippet(snippetMap interface{}) (*gofastly.CreateSnippetInput, error) {
	df := snippetMap.(map[string]interface{})
	opts := gofastly.CreateSnippetInput{
		Name:     df["name"].(string),
		Content:  df["content"].(string),
		Priority: df["priority"].(int),
	}

	snippetType := strings.ToLower(df["type"].(string))
	switch snippetType {
	case "init":
		opts.Type = gofastly.SnippetTypeInit
	case "recv":
		opts.Type = gofastly.SnippetTypeRecv
	case "hash":
		opts.Type = gofastly.SnippetTypeHash
	case "hit":
		opts.Type = gofastly.SnippetTypeHit
	case "miss":
		opts.Type = gofastly.SnippetTypeMiss
	case "pass":
		opts.Type = gofastly.SnippetTypePass
	case "fetch":
		opts.Type = gofastly.SnippetTypeFetch
	case "error":
		opts.Type = gofastly.SnippetTypeError
	case "deliver":
		opts.Type = gofastly.SnippetTypeDeliver
	case "log":
		opts.Type = gofastly.SnippetTypeLog
	case "none":
		opts.Type = gofastly.SnippetTypeNone
	}

	return &opts, nil
}

func buildDynamicSnippet(dynamicSnippetMap interface{}) (*gofastly.CreateSnippetInput, error) {
	df := dynamicSnippetMap.(map[string]interface{})
	opts := gofastly.CreateSnippetInput{
		Name:     df["name"].(string),
		Priority: df["priority"].(int),
		Dynamic:  1,
	}

	snippetType := strings.ToLower(df["type"].(string))
	switch snippetType {
	case "init":
		opts.Type = gofastly.SnippetTypeInit
	case "recv":
		opts.Type = gofastly.SnippetTypeRecv
	case "hash":
		opts.Type = gofastly.SnippetTypeHash
	case "hit":
		opts.Type = gofastly.SnippetTypeHit
	case "miss":
		opts.Type = gofastly.SnippetTypeMiss
	case "pass":
		opts.Type = gofastly.SnippetTypePass
	case "fetch":
		opts.Type = gofastly.SnippetTypeFetch
	case "error":
		opts.Type = gofastly.SnippetTypeError
	case "deliver":
		opts.Type = gofastly.SnippetTypeDeliver
	case "log":
		opts.Type = gofastly.SnippetTypeLog
	case "none":
		opts.Type = gofastly.SnippetTypeNone
	}

	return &opts, nil
}

func flattenSnippets(snippetList []*gofastly.Snippet) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, snippet := range snippetList {
		// Skip dynamic snippets
		if snippet.Dynamic == 1 {
			continue
		}

		// Convert VCLs to a map for saving to state.
		snippetMap := map[string]interface{}{
			"name":     snippet.Name,
			"type":     snippet.Type,
			"priority": int(snippet.Priority),
			"content":  snippet.Content,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range snippetMap {
			if v == "" {
				delete(snippetMap, k)
			}
		}

		sl = append(sl, snippetMap)
	}

	return sl
}

func flattenDynamicSnippets(dynamicSnippetList []*gofastly.Snippet) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, dynamicSnippet := range dynamicSnippetList {
		// Skip non-dynamic snippets
		if dynamicSnippet.Dynamic == 0 {
			continue
		}

		// Convert VCLs to a map for saving to state.
		dynamicSnippetMap := map[string]interface{}{
			"snippet_id": dynamicSnippet.ID,
			"name":       dynamicSnippet.Name,
			"type":       dynamicSnippet.Type,
			"priority":   int(dynamicSnippet.Priority),
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range dynamicSnippetMap {
			if v == "" {
				delete(dynamicSnippetMap, k)
			}
		}

		sl = append(sl, dynamicSnippetMap)
	}

	return sl
}

func buildDictionary(dictMap interface{}) (*gofastly.CreateDictionaryInput, error) {
	df := dictMap.(map[string]interface{})
	opts := gofastly.CreateDictionaryInput{
		Name:      df["name"].(string),
		WriteOnly: gofastly.CBool(df["write_only"].(bool)),
	}

	return &opts, nil
}

func flattenDictionaries(dictList []*gofastly.Dictionary) []map[string]interface{} {
	var dl []map[string]interface{}
	for _, currentDict := range dictList {

		dictMapString := map[string]interface{}{
			"dictionary_id": currentDict.ID,
			"name":          currentDict.Name,
			"write_only":    currentDict.WriteOnly,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range dictMapString {
			if v == "" {
				delete(dictMapString, k)
			}
		}

		dl = append(dl, dictMapString)
	}

	return dl
}

func validateVCLs(d *schema.ResourceData) error {
	// TODO: this would be nice to move into a resource/collection validation function, once that is available
	// (see https://github.com/hashicorp/terraform/pull/4348 and https://github.com/hashicorp/terraform/pull/6508)
	vcls, exists := d.GetOk("vcl")
	if !exists {
		return nil
	}

	numberOfMainVCLs, numberOfIncludeVCLs := 0, 0
	for _, vclElem := range vcls.(*schema.Set).List() {
		vcl := vclElem.(map[string]interface{})
		if mainVal, hasMain := vcl["main"]; hasMain && mainVal.(bool) {
			numberOfMainVCLs++
		} else {
			numberOfIncludeVCLs++
		}
	}
	if numberOfMainVCLs == 0 && numberOfIncludeVCLs > 0 {
		return errors.New("if you include VCL configurations, one of them should have main = true")
	}
	if numberOfMainVCLs > 1 {
		return errors.New("you cannot have more than one VCL configuration with main = true")
	}
	return nil
}
