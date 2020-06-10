package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
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


func processVCL(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	// Note: as above with Gzip and S3 logging, we don't utilize the PUT
	// endpoint to update a VCL, we simply destroy it and create a new one.
	oldVCLVal, newVCLVal := d.GetChange("vcl")
	if oldVCLVal == nil {
		oldVCLVal = new(schema.Set)
	}
	if newVCLVal == nil {
		newVCLVal = new(schema.Set)
	}

	oldVCLSet := oldVCLVal.(*schema.Set)
	newVCLSet := newVCLVal.(*schema.Set)

	remove := oldVCLSet.Difference(newVCLSet).List()
	add := newVCLSet.Difference(oldVCLSet).List()

	// Delete removed VCL configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteVCLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Removal opts: %#v", opts)
		err := conn.DeleteVCL(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	// POST new VCL configurations
	for _, dRaw := range add {
		df := dRaw.(map[string]interface{})
		opts := gofastly.CreateVCLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
			Content: df["content"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Addition opts: %#v", opts)
		_, err := conn.CreateVCL(&opts)
		if err != nil {
			return err
		}

		// if this new VCL is the main
		if df["main"].(bool) {
			opts := gofastly.ActivateVCLInput{
				Service: d.Id(),
				Version: latestVersion,
				Name:    df["name"].(string),
			}
			log.Printf("[DEBUG] Fastly VCL activation opts: %#v", opts)
			_, err := conn.ActivateVCL(&opts)
			if err != nil {
				return err
			}

		}
	}
	return nil
}
