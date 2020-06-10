package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var conditionSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"statement": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The statement used to determine if the condition is met",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of the condition, either `REQUEST`, `RESPONSE`, or `CACHE`",
				ValidateFunc: validateConditionType(),
			},
		},
	},
}