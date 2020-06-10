package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var domainSchema = &schema.Schema{
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
}