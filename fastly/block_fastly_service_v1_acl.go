package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

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


func processACL(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	oldACLVal, newACLVal := d.GetChange("acl")
	if oldACLVal == nil {
		oldACLVal = new(schema.Set)
	}
	if newACLVal == nil {
		newACLVal = new(schema.Set)
	}

	oldACLSet := oldACLVal.(*schema.Set)
	newACLSet := newACLVal.(*schema.Set)

	remove := oldACLSet.Difference(newACLSet).List()
	add := newACLSet.Difference(oldACLSet).List()

	// Delete removed ACL configurations
	for _, vRaw := range remove {
		val := vRaw.(map[string]interface{})
		opts := gofastly.DeleteACLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    val["name"].(string),
		}

		log.Printf("[DEBUG] Fastly ACL removal opts: %#v", opts)
		err := conn.DeleteACL(&opts)

		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new ACL configurations
	for _, vRaw := range add {
		val := vRaw.(map[string]interface{})
		opts := gofastly.CreateACLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    val["name"].(string),
		}

		log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
		_, err := conn.CreateACL(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}
