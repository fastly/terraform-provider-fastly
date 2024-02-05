package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CacheSettingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type CacheSettingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceCacheSetting returns a new resource.
func NewServiceCacheSetting(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&CacheSettingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "cache_setting",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *CacheSettingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *CacheSettingServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: `One of cache, pass, or restart, as defined on Fastly's documentation under "[Caching action descriptions](https://docs.fastly.com/en/guides/controlling-caching#caching-action-descriptions)"`,
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` used to test whether this settings object should be used. This `condition` must be of type `CACHE`",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name for this Cache Setting. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"stale_ttl": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: `Max "Time To Live" for stale (unreachable) objects`,
				},
				"ttl": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The Time-To-Live (TTL) for the object",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *CacheSettingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := &gofastly.CreateCacheSettingInput{
		Name:           gofastly.ToPointer(resource["name"].(string)),
		StaleTTL:       gofastly.ToPointer(resource["stale_ttl"].(int)),
		CacheCondition: gofastly.ToPointer(resource["cache_condition"].(string)),
	}

	if v, ok := resource["ttl"]; ok {
		opts.TTL = gofastly.ToPointer(v.(int))
	}

	act := strings.ToLower(resource["action"].(string))
	switch act {
	case "cache":
		opts.Action = gofastly.ToPointer(gofastly.CacheSettingActionCache)
	case "pass":
		opts.Action = gofastly.ToPointer(gofastly.CacheSettingActionPass)
	case "restart":
		opts.Action = gofastly.ToPointer(gofastly.CacheSettingActionRestart)
	}
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Fastly Cache Settings Addition opts: %#v", opts)
	if _, err := conn.CreateCacheSetting(opts); err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *CacheSettingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Cache Settings for (%s)", d.Id())
		remoteState, err := conn.ListCacheSettings(&gofastly.ListCacheSettingsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Cache Settings for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		csl := flattenCacheSettings(remoteState)

		if err := d.Set(h.GetKey(), csl); err != nil {
			log.Printf("[WARN] Error setting Cache Settings for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *CacheSettingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateCacheSettingInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["action"]; ok {
		opts.Action = gofastly.ToPointer(gofastly.CacheSettingAction(v.(string)))
	}
	if v, ok := modified["ttl"]; ok {
		opts.TTL = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["stale_ttl"]; ok {
		opts.StaleTTL = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Cache Setting Opts: %#v", opts)
	_, err := conn.UpdateCacheSetting(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *CacheSettingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteCacheSettingInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
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
	return nil
}

// flattenCacheSettings models data into format suitable for saving to Terraform state.
func flattenCacheSettings(remoteState []*gofastly.CacheSetting) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Action != nil {
			data["action"] = *resource.Action
		}
		if resource.CacheCondition != nil {
			data["cache_condition"] = *resource.CacheCondition
		}
		if resource.StaleTTL != nil {
			data["stale_ttl"] = *resource.StaleTTL
		}
		if resource.TTL != nil {
			data["ttl"] = *resource.TTL
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
