package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var healthcheckSchema = &schema.Schema{
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
}