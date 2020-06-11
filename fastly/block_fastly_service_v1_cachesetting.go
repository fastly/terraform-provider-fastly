package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type CacheSettingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceCacheSetting() ServiceAttributeDefinition {
	return &CacheSettingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:    "cache_setting",
		},
	}
}


func (h *CacheSettingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
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

func (h *CacheSettingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Cache Settings for (%s)", d.Id())
	cslList, err := conn.ListCacheSettings(&gofastly.ListCacheSettingsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Cache Settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	csl := flattenCacheSettings(cslList)

	if err := d.Set("cache_setting", csl); err != nil {
		log.Printf("[WARN] Error setting Cache Settings for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *CacheSettingServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
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
	return nil
}


func buildCacheSetting(cacheMap interface{}) (*gofastly.CreateCacheSettingInput, error) {
	df := cacheMap.(map[string]interface{})
	opts := gofastly.CreateCacheSettingInput{
		Name:           df["name"].(string),
		StaleTTL:       uint(df["stale_ttl"].(int)),
		CacheCondition: df["cache_condition"].(string),
	}

	if v, ok := df["ttl"]; ok {
		opts.TTL = uint(v.(int))
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "cache":
		opts.Action = gofastly.CacheSettingActionCache
	case "pass":
		opts.Action = gofastly.CacheSettingActionPass
	case "restart":
		opts.Action = gofastly.CacheSettingActionRestart
	}

	return &opts, nil
}

func flattenCacheSettings(csList []*gofastly.CacheSetting) []map[string]interface{} {
	var csl []map[string]interface{}
	for _, cl := range csList {
		// Convert Cache Settings to a map for saving to state.
		clMap := map[string]interface{}{
			"name":            cl.Name,
			"action":          cl.Action,
			"cache_condition": cl.CacheCondition,
			"stale_ttl":       cl.StaleTTL,
			"ttl":             cl.TTL,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range clMap {
			if v == "" {
				delete(clMap, k)
			}
		}

		csl = append(csl, clMap)
	}

	return csl
}
