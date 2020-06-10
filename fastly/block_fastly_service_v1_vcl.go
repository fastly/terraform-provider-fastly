package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var vclSchema = &schema.Schema{
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
}