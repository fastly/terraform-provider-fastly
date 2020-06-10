package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"


var blobstorageloggingSchema = &schema.Schema{
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
}