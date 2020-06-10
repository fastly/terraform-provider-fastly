package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var headerSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name to refer to this Header object",
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of set, append, delete, regex, or regex_repeat",
				ValidateFunc: validateHeaderAction(),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type to manipulate: request, fetch, cache, response",
				ValidateFunc: validateHeaderType(),
			},
			"destination": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Header this affects",
			},
			// Optional fields, defaults where they exist
			"ignore_if_set": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Don't add the header if it is already. (Only applies to 'set' action.). Default `false`",
			},
			"source": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Variable to be used as a source for the header content (Does not apply to 'delete' action.)",
			},
			"regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Regular expression to use (Only applies to 'regex' and 'regex_repeat' actions.)",
			},
			"substitution": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Value to substitute in place of regular expression. (Only applies to 'regex' and 'regex_repeat'.)",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Lower priorities execute first. (Default: 100.)",
			},
			"request_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Optional name of a request condition to apply.",
			},
			"cache_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Optional name of a cache condition to apply.",
			},
			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Optional name of a response condition to apply.",
			},
		},
	},
}