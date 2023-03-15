package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// WAFServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type WAFServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceWAF returns a new resource.
func NewServiceWAF(sa ServiceMetadata) ServiceAttributeDefinition {
	return &WAFServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "waf",
			serviceMetadata: sa,
		},
	}
}

// Register add the attribute to the resource schema.
func (h *WAFServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "A flag used to completely disable a Web Application Firewall. This is intended to only be used in an emergency",
				},
				"prefetch_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The `condition` to determine which requests will be run past your Fastly WAF. This `condition` must be of type `PREFETCH`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
				"response_object": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the response object used by the Web Application Firewall",
				},
				"waf_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the WAF",
				},
			},
		},
	}

	return nil
}

// Process creates or updates the attribute against the Fastly API.
func (h *WAFServiceAttributeHandler) Process(_ context.Context, d *schema.ResourceData, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	oldWAFVal, newWAFVal := d.GetChange(h.GetKey())

	if len(newWAFVal.([]any)) == 1 {
		wf := newWAFVal.([]any)[0].(map[string]any)

		var err error
		if wafExists(conn, serviceID, serviceVersion, wf["waf_id"].(string)) {
			opts := buildUpdateWAF(d, wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF update opts: %#v", opts)
			_, err = conn.UpdateWAF(opts)
		} else {
			opts := buildCreateWAF(wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF Addition opts: %#v", opts)

			_, err = conn.CreateWAF(opts)
		}
		if err != nil {
			return err
		}
	} else if len(oldWAFVal.([]any)) > 0 {
		wf := oldWAFVal.([]any)[0].(map[string]any)

		opts := buildDeleteWAF(wf, serviceVersion)
		log.Printf("[DEBUG] Fastly WAF Removal opts: %#v", opts)
		err := conn.DeleteWAF(opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (h *WAFServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).([]any)

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing WAFs for (%s)", d.Id())
		remoteState, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: d.Id(),
			FilterVersion: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAFs for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		waf := flattenWAFs(remoteState.Items)

		if err := d.Set("waf", waf); err != nil {
			log.Printf("[WARN] Error setting waf for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

func wafExists(conn *gofastly.Client, s string, v int, id string) bool {
	_, err := conn.GetWAF(&gofastly.GetWAFInput{
		ServiceID:      s,
		ServiceVersion: v,
		ID:             id,
	})
	return err == nil
}

// flattenWAFs models data into format suitable for saving to Terraform state.
func flattenWAFs(remoteState []*gofastly.WAF) []map[string]any {
	var result []map[string]any
	if len(remoteState) == 0 {
		return result
	}

	w := remoteState[0]
	data := map[string]any{
		"waf_id":             w.ID,
		"response_object":    w.Response,
		"prefetch_condition": w.PrefetchCondition,
	}

	// prune any empty values that come from the default string value in structs
	for k, v := range data {
		if v == "" {
			delete(data, k)
		}
	}
	return append(result, data)
}

func buildCreateWAF(waf any, serviceID string, serviceVersion int) *gofastly.CreateWAFInput {
	resource := waf.(map[string]any)

	opts := gofastly.CreateWAFInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		ID:                resource["waf_id"].(string),
		PrefetchCondition: resource["prefetch_condition"].(string),
		Response:          resource["response_object"].(string),
	}
	return &opts
}

func buildDeleteWAF(waf any, serviceVersion int) *gofastly.DeleteWAFInput {
	resource := waf.(map[string]any)

	opts := gofastly.DeleteWAFInput{
		ID:             resource["waf_id"].(string),
		ServiceVersion: serviceVersion,
	}
	return &opts
}

func buildUpdateWAF(d *schema.ResourceData, wafMap any, serviceID string, serviceVersion int) *gofastly.UpdateWAFInput {
	resource := wafMap.(map[string]any)

	input := gofastly.UpdateWAFInput{
		ServiceID:      gofastly.String(serviceID),
		ServiceVersion: gofastly.Int(serviceVersion),
		ID:             resource["waf_id"].(string),
	}

	// NOTE: to access the WAF data we need to link to a specific list index.
	// This is because the schema defines the service as being of TypeList.
	//
	// Although there should only ever be one service (hence MaxItems: 1) we are
	// unable to change the schema to a TypeMap as that would constrain the map's
	// value to a single type (e.g. TypeString, TypeBool, TypeInt, or TypeFloat).

	if v, ok := d.GetOk("waf.0.prefetch_condition"); ok {
		input.PrefetchCondition = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("waf.0.response_object"); ok {
		input.Response = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("waf.0.disabled"); ok {
		input.Disabled = gofastly.Bool(v.(bool))
	}

	return &input
}
