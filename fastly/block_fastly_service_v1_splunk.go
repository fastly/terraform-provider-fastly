package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var  splunkSchema = &schema.Schema{
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
}