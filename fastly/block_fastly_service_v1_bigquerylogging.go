package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var bigqueryloggingSchema = &schema.Schema{
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
}