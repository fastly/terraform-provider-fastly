package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var gzipSchema = &schema.Schema{
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
}