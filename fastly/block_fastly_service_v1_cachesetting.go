package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var cachesettingSchema = &schema.Schema{
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
}

