package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var logentriesSchema = &schema.Schema{
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
}