package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// RequestSettingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type RequestSettingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceRequestSetting returns a new resource.
func NewServiceRequestSetting(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&RequestSettingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "request_setting",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *RequestSettingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *RequestSettingServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
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
				"default_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Sets the host header",
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
				// TODO: Although Fastly API has been exposing this parameter over years
				// it turned out that setting this parameter does nothing. We should remove this attribute in v3.0.0
				"geo_headers": {
					Type:        schema.TypeBool,
					Optional:    true,
					Deprecated:  "'geo_headers' attribute has been deprecated and will be removed in the next major version release",
					Description: "Injects Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region into the request headers",
				},
				"hash_keys": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Comma separated list of varnish request object fields that should be in the hash key",
				},
				"max_stale_age": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "How old an object is allowed to be to serve `stale-if-error` or `stale-while-revalidate`, in seconds",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name to refer to this Request Setting. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to determine if this request setting should be applied (should be unique across multiple instances of `request_setting`)",
				},
				"timer_support": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Injects the X-Timer info into the request for viewing origin fetch durations",
				},
				"xff": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "append",
					Description: "X-Forwarded-For, should be `clear`, `leave`, `append`, `append_all`, or `overwrite`. Default `append`",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *RequestSettingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts, err := buildRequestSetting(resource)
	if err != nil {
		log.Printf("[DEBUG] Error building Request Setting: %s", err)
		return err
	}
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
	_, err = conn.CreateRequestSetting(opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *RequestSettingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
		remoteState, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Request Settings for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		rl := flattenRequestSettings(remoteState)

		if err := d.Set(h.GetKey(), rl); err != nil {
			log.Printf("[WARN] Error setting Request Settings for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *RequestSettingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateRequestSettingInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	if v, ok := modified["force_miss"]; ok {
		opts.ForceMiss = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["force_ssl"]; ok {
		opts.ForceSSL = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["action"]; ok {
		switch strings.ToLower(v.(string)) {
		case "lookup":
			opts.Action = gofastly.RequestSettingActionLookup
		case "pass":
			opts.Action = gofastly.RequestSettingActionPass
		}
	}
	if v, ok := modified["bypass_busy_wait"]; ok {
		opts.BypassBusyWait = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["max_stale_age"]; ok {
		opts.MaxStaleAge = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["hash_keys"]; ok {
		opts.HashKeys = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["xff"]; ok {
		switch strings.ToLower(v.(string)) {
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
	}
	if v, ok := modified["timer_support"]; ok {
		opts.TimerSupport = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["geo_headers"]; ok {
		opts.GeoHeaders = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["default_host"]; ok {
		opts.DefaultHost = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_condition"]; ok {
		opts.RequestCondition = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Request Settings Opts: %#v", opts)
	_, err := conn.UpdateRequestSetting(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *RequestSettingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteRequestSettingInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
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
	return nil
}

// flattenRequestSettings models data into format suitable for saving to Terraform state.
func flattenRequestSettings(remoteState []*gofastly.RequestSetting) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":              resource.Name,
			"max_stale_age":     resource.MaxStaleAge,
			"force_miss":        resource.ForceMiss,
			"force_ssl":         resource.ForceSSL,
			"action":            resource.Action,
			"bypass_busy_wait":  resource.BypassBusyWait,
			"hash_keys":         resource.HashKeys,
			"xff":               resource.XForwardedFor,
			"timer_support":     resource.TimerSupport,
			"geo_headers":       resource.GeoHeaders,
			"default_host":      resource.DefaultHost,
			"request_condition": resource.RequestCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func buildRequestSetting(requestSettingMap any) (*gofastly.CreateRequestSettingInput, error) {
	resource := requestSettingMap.(map[string]any)
	opts := gofastly.CreateRequestSettingInput{
		Name:           gofastly.ToPointer(resource["name"].(string)),
		MaxStaleAge:    gofastly.ToPointer(resource["max_stale_age"].(int)),
		ForceMiss:      gofastly.ToPointer(gofastly.Compatibool(resource["force_miss"].(bool))),
		ForceSSL:       gofastly.ToPointer(gofastly.Compatibool(resource["force_ssl"].(bool))),
		BypassBusyWait: gofastly.ToPointer(gofastly.Compatibool(resource["bypass_busy_wait"].(bool))),
		HashKeys:       gofastly.ToPointer(resource["hash_keys"].(string)),
		TimerSupport:   gofastly.ToPointer(gofastly.Compatibool(resource["timer_support"].(bool))),
		GeoHeaders:     gofastly.ToPointer(gofastly.Compatibool(resource["geo_headers"].(bool))),
		DefaultHost:    gofastly.ToPointer(resource["default_host"].(string)),
	}

	if v := resource["request_condition"].(string); v != "" {
		opts.RequestCondition = gofastly.ToPointer(v)
	}

	act := strings.ToLower(resource["action"].(string))
	switch act {
	case "lookup":
		opts.Action = gofastly.ToPointer(gofastly.RequestSettingActionLookup)
	case "pass":
		opts.Action = gofastly.ToPointer(gofastly.RequestSettingActionPass)
	}

	xff := strings.ToLower(resource["xff"].(string))
	switch xff {
	case "clear":
		opts.XForwardedFor = gofastly.ToPointer(gofastly.RequestSettingXFFClear)
	case "leave":
		opts.XForwardedFor = gofastly.ToPointer(gofastly.RequestSettingXFFLeave)
	case "append":
		opts.XForwardedFor = gofastly.ToPointer(gofastly.RequestSettingXFFAppend)
	case "append_all":
		opts.XForwardedFor = gofastly.ToPointer(gofastly.RequestSettingXFFAppendAll)
	case "overwrite":
		opts.XForwardedFor = gofastly.ToPointer(gofastly.RequestSettingXFFOverwrite)
	}

	return &opts, nil
}
