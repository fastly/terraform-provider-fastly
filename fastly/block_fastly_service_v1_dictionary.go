package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

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


func processDictionary(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	oldDictVal, newDictVal := d.GetChange("dictionary")

	if oldDictVal == nil {
		oldDictVal = new(schema.Set)
	}
	if newDictVal == nil {
		newDictVal = new(schema.Set)
	}

	oldDictSet := oldDictVal.(*schema.Set)
	newDictSet := newDictVal.(*schema.Set)

	remove := oldDictSet.Difference(newDictSet).List()
	add := newDictSet.Difference(oldDictSet).List()

	// Delete removed dictionary configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteDictionaryInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Dictionary Removal opts: %#v", opts)
		err := conn.DeleteDictionary(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new dictionary configurations
	for _, dRaw := range add {
		opts, err := buildDictionary(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Dicitionary: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly Dictionary Addition opts: %#v", opts)
		_, err = conn.CreateDictionary(opts)
		if err != nil {
			return err
		}
	}
	return nil
}