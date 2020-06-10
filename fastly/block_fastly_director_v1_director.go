package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

var directorSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name to refer to this director",
			},
			"backends": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of backends associated with this director",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			// optional fields
			"capacity": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Load balancing weight for the backends",
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"shield": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Selected POP to serve as a 'shield' for origin servers.",
			},
			"quorum": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      75,
				Description:  "Percentage of capacity that needs to be up for the director itself to be considered up",
				ValidateFunc: validateDirectorQuorum(),
			},
			"type": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Type of load balance group to use. Integer, 1 to 4. Values: 1 (random), 3 (hash), 4 (client)",
				ValidateFunc: validateDirectorType(),
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "How many backends to search if it fails",
			},
		},
	},
}