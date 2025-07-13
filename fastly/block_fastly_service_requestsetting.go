package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
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
func (h *RequestSettingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := buildRequestSetting(resource)
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
	_, err := conn.CreateRequestSetting(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *RequestSettingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
		remoteState, err := conn.ListRequestSettings(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListRequestSettingsInput{
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
func (h *RequestSettingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
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
			opts.Action = gofastly.ToPointer(gofastly.RequestSettingActionLookup)
		case "pass":
			opts.Action = gofastly.ToPointer(gofastly.RequestSettingActionPass)
		default:
			opts.Action = gofastly.ToPointer(gofastly.RequestSettingActionUnset)
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
	}
	if v, ok := modified["timer_support"]; ok {
		opts.TimerSupport = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["default_host"]; ok {
		opts.DefaultHost = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_condition"]; ok {
		opts.RequestCondition = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Request Settings Opts: %#v", opts)
	_, err := conn.UpdateRequestSetting(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *RequestSettingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteRequestSettingInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Request Setting removal opts: %#v", opts)
	err := conn.DeleteRequestSetting(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
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
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.MaxStaleAge != nil {
			data["max_stale_age"] = *resource.MaxStaleAge
		}
		if resource.ForceMiss != nil {
			data["force_miss"] = *resource.ForceMiss
		}
		if resource.ForceSSL != nil {
			data["force_ssl"] = *resource.ForceSSL
		}
		if resource.Action != nil {
			data["action"] = *resource.Action
		}
		if resource.BypassBusyWait != nil {
			data["bypass_busy_wait"] = *resource.BypassBusyWait
		}
		if resource.HashKeys != nil {
			data["hash_keys"] = *resource.HashKeys
		}
		if resource.XForwardedFor != nil {
			data["xff"] = *resource.XForwardedFor
		}
		if resource.TimerSupport != nil {
			data["timer_support"] = *resource.TimerSupport
		}
		if resource.DefaultHost != nil {
			data["default_host"] = *resource.DefaultHost
		}
		if resource.RequestCondition != nil {
			data["request_condition"] = *resource.RequestCondition
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

func buildRequestSetting(requestSettingMap any) *gofastly.CreateRequestSettingInput {
	resource := requestSettingMap.(map[string]any)
	opts := gofastly.CreateRequestSettingInput{
		BypassBusyWait: gofastly.ToPointer(gofastly.Compatibool(resource["bypass_busy_wait"].(bool))),
		DefaultHost:    gofastly.ToPointer(resource["default_host"].(string)),
		ForceMiss:      gofastly.ToPointer(gofastly.Compatibool(resource["force_miss"].(bool))),
		ForceSSL:       gofastly.ToPointer(gofastly.Compatibool(resource["force_ssl"].(bool))),
		HashKeys:       gofastly.ToPointer(resource["hash_keys"].(string)),
		MaxStaleAge:    gofastly.ToPointer(resource["max_stale_age"].(int)),
		Name:           gofastly.ToPointer(resource["name"].(string)),
		TimerSupport:   gofastly.ToPointer(gofastly.Compatibool(resource["timer_support"].(bool))),
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

	return &opts
}
