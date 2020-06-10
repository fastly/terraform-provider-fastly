package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var cachesettingSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name to refer to this Cache Setting",
			},
			"action": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Action to take",
			},
			// optional
			"cache_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition to check if this Cache Setting applies",
			},
			"stale_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max 'Time To Live' for stale (unreachable) objects.",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The 'Time To Live' for the object",
			},
		},
	},
}


func processCacheSetting(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	oc, nc := d.GetChange("cache_setting")
	if oc == nil {
		oc = new(schema.Set)
	}
	if nc == nil {
		nc = new(schema.Set)
	}

	ocs := oc.(*schema.Set)
	ncs := nc.(*schema.Set)

	remove := ocs.Difference(ncs).List()
	add := ncs.Difference(ocs).List()

	// Delete removed Cache Settings
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteCacheSettingInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Cache Settings removal opts: %#v", opts)
		err := conn.DeleteCacheSetting(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new Cache Settings
	for _, dRaw := range add {
		opts, err := buildCacheSetting(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Cache Setting: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly Cache Settings Addition opts: %#v", opts)
		_, err = conn.CreateCacheSetting(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

