package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var requestsettingSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this Request Setting",
			},
			// Optional fields
			"request_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a request condition to apply. If there is no condition this setting will always be applied.",
			},
			"max_stale_age": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How old an object is allowed to be, in seconds. Default `60`",
			},
			"force_miss": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Force a cache miss for the request",
			},
			"force_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Forces the request use SSL",
			},
			"action": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allows you to terminate request handling and immediately perform an action",
			},
			"bypass_busy_wait": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable collapsed forwarding",
			},
			"hash_keys": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comma separated list of varnish request object fields that should be in the hash key",
			},
			"xff": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "append",
				Description: "X-Forwarded-For options",
			},
			"timer_support": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Injects the X-Timer info into the request",
			},
			"geo_headers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Inject Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region",
			},
			"default_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "the host header",
			},
		},
	},
}


func processRequestSetting(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	os, ns := d.GetChange("request_setting")
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	ors := os.(*schema.Set)
	nrs := ns.(*schema.Set)
	removeRequestSettings := ors.Difference(nrs).List()
	addRequestSettings := nrs.Difference(ors).List()

	// DELETE old Request Settings configurations
	for _, sRaw := range removeRequestSettings {
		sf := sRaw.(map[string]interface{})
		opts := gofastly.DeleteRequestSettingInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    sf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Request Setting removal opts: %#v", opts)
		err := conn.DeleteRequestSetting(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Request Setting
	for _, sRaw := range addRequestSettings {
		opts, err := buildRequestSetting(sRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Requset Setting: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
		_, err = conn.CreateRequestSetting(opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func readRequestSetting(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
	rsList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Request Settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	rl := flattenRequestSettings(rsList)

	if err := d.Set("request_setting", rl); err != nil {
		log.Printf("[WARN] Error setting Request Settings for (%s): %s", d.Id(), err)
	}
	return nil
}
