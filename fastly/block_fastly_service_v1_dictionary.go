package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var dictionarySchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this Dictionary",
			},
			// Optional fields
			"dictionary_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Generated dictionary ID",
			},
			"write_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Determines if items in the dictionary are readable or not",
			},
		},
	},
}