package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var aclSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this ACL",
			},
			// Optional fields
			"acl_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Generated acl id",
			},
		},
	},
}
