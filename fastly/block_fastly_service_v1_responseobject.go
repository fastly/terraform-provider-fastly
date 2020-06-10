package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var responseobjectSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this request object",
			},
			// Optional fields
			"status": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     200,
				Description: "The HTTP Status Code of the object",
			},
			"response": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OK",
				Description: "The HTTP Response of the object",
			},
			"content": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The content to deliver for the response object",
			},
			"content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The MIME type of the content",
			},
			"request_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of the condition to be checked during the request phase to see if the object should be delivered",
			},
			"cache_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of the condition checked after we have retrieved an object. If the condition passes then deliver this Request Object instead.",
			},
		},
	},
}