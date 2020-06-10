package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var syslogSchema = &schema.Schema{
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
}

