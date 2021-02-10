package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type RequestSettingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceRequestSetting(sa ServiceMetadata) ServiceAttributeDefinition {
	return &RequestSettingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "request_setting",
			serviceMetadata: sa,
		},
	}
}

func (h *RequestSettingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	ors := os.(*schema.Set)
	nrs := ns.(*schema.Set)

	setDiff := NewSetDiff(func(reqsettings interface{}) (interface{}, error) {
		// Use the request settings name as the key
		return reqsettings.(map[string]interface{})["name"], nil
	})

	diffResult, err := setDiff.Diff(ors, nrs)
	if err != nil {
		return err
	}

	// DELETE old Request Settings configurations
	for _, sRaw := range diffResult.Deleted {
		sf := sRaw.(map[string]interface{})
		opts := gofastly.DeleteRequestSettingInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           sf["name"].(string),
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
	for _, sRaw := range diffResult.Added {
		opts, err := buildRequestSetting(sRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Requset Setting: %s", err)
			return err
		}
		opts.ServiceID = d.Id()
		opts.ServiceVersion = latestVersion

		log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
		_, err = conn.CreateRequestSetting(opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *RequestSettingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
	rsList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Request Settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	rl := flattenRequestSettings(rsList)

	if err := d.Set(h.GetKey(), rl); err != nil {
		log.Printf("[WARN] Error setting Request Settings for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *RequestSettingServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
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
					Description: "Name of already defined `condition` to determine if this request setting should be applied",
				},
				"max_stale_age": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "How old an object is allowed to be to serve `stale-if-error` or `stale-while-revalidate`, in seconds",
				},
				"force_miss": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Force a cache miss for the request. If specified, can be `true` or `false`",
				},
				"force_ssl": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Forces the request to use SSL (Redirects a non-SSL request to SSL)",
				},
				"action": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Allows you to terminate request handling and immediately perform an action. When set it can be `lookup` or `pass` (Ignore the cache completely)",
				},
				"bypass_busy_wait": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disable collapsed forwarding, so you don't wait for other objects to origin",
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
					Description: "X-Forwarded-For, should be `clear`, `leave`, `append`, `append_all`, or `overwrite`. Default `append`",
				},
				"timer_support": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Injects the X-Timer info into the request for viewing origin fetch durations",
				},
				"geo_headers": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Injects Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region into the request headers",
				},
				"default_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Sets the host header",
				},
			},
		},
	}
	return nil
}

func flattenRequestSettings(rsList []*gofastly.RequestSetting) []map[string]interface{} {
	var rl []map[string]interface{}
	for _, r := range rsList {
		// Convert Request Settings to a map for saving to state.
		nrs := map[string]interface{}{
			"name":              r.Name,
			"max_stale_age":     r.MaxStaleAge,
			"force_miss":        r.ForceMiss,
			"force_ssl":         r.ForceSSL,
			"action":            r.Action,
			"bypass_busy_wait":  r.BypassBusyWait,
			"hash_keys":         r.HashKeys,
			"xff":               r.XForwardedFor,
			"timer_support":     r.TimerSupport,
			"geo_headers":       r.GeoHeaders,
			"default_host":      r.DefaultHost,
			"request_condition": r.RequestCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nrs {
			if v == "" {
				delete(nrs, k)
			}
		}

		rl = append(rl, nrs)
	}

	return rl
}

func buildRequestSetting(requestSettingMap interface{}) (*gofastly.CreateRequestSettingInput, error) {
	df := requestSettingMap.(map[string]interface{})
	opts := gofastly.CreateRequestSettingInput{
		Name:             df["name"].(string),
		MaxStaleAge:      uint(df["max_stale_age"].(int)),
		ForceMiss:        gofastly.Compatibool(df["force_miss"].(bool)),
		ForceSSL:         gofastly.Compatibool(df["force_ssl"].(bool)),
		BypassBusyWait:   gofastly.Compatibool(df["bypass_busy_wait"].(bool)),
		HashKeys:         df["hash_keys"].(string),
		TimerSupport:     gofastly.Compatibool(df["timer_support"].(bool)),
		GeoHeaders:       gofastly.Compatibool(df["geo_headers"].(bool)),
		DefaultHost:      df["default_host"].(string),
		RequestCondition: df["request_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "lookup":
		opts.Action = gofastly.RequestSettingActionLookup
	case "pass":
		opts.Action = gofastly.RequestSettingActionPass
	}

	xff := strings.ToLower(df["xff"].(string))
	switch xff {
	case "clear":
		opts.XForwardedFor = gofastly.RequestSettingXFFClear
	case "leave":
		opts.XForwardedFor = gofastly.RequestSettingXFFLeave
	case "append":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppend
	case "append_all":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppendAll
	case "overwrite":
		opts.XForwardedFor = gofastly.RequestSettingXFFOverwrite
	}

	return &opts, nil
}
