package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HoneycombServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type HoneycombServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingHoneycomb returns a new resource.
func NewServiceLoggingHoneycomb(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&HoneycombServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_honeycomb",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *HoneycombServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *HoneycombServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"dataset": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Honeycomb Dataset you want to log to",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Honeycomb logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The Write Key from the Account page of your Honeycomb account",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting. Your log must produce valid JSON that Honeycomb can ingest.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *HoneycombServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Honeycomb logging addition opts: %#v", opts)

	return createHoneycomb(conn, opts)
}

// Read refreshes the resource.
func (h *HoneycombServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Honeycomb logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListHoneycombs(&gofastly.ListHoneycombsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Honeycomb logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenHoneycomb(remoteState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting Honeycomb logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *HoneycombServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateHoneycombInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["dataset"]; ok {
		opts.Dataset = gofastly.String(v.(string))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Honeycomb Opts: %#v", opts)
	_, err := conn.UpdateHoneycomb(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HoneycombServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Honeycomb logging endpoint removal opts: %#v", opts)

	return deleteHoneycomb(conn, opts)
}

func createHoneycomb(conn *gofastly.Client, i *gofastly.CreateHoneycombInput) error {
	_, err := conn.CreateHoneycomb(i)
	return err
}

func deleteHoneycomb(conn *gofastly.Client, i *gofastly.DeleteHoneycombInput) error {
	err := conn.DeleteHoneycomb(i)

	errRes, ok := err.(*gofastly.HTTPError)
	if !ok {
		return err
	}

	// 404 response codes don't result in an error propagating because a 404 could
	// indicate that a resource was deleted elsewhere.
	if !errRes.IsNotFound() {
		return err
	}

	return nil
}

// flattenHoneycomb models data into format suitable for saving to Terraform state.
func flattenHoneycomb(remoteState []*gofastly.Honeycomb) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"token":              resource.Token,
			"dataset":            resource.Dataset,
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"placement":          resource.Placement,
			"response_condition": resource.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func (h *HoneycombServiceAttributeHandler) buildCreate(honeycombMap any, serviceID string, serviceVersion int) *gofastly.CreateHoneycombInput {
	resource := honeycombMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateHoneycombInput{
		Dataset:        gofastly.String(resource["dataset"].(string)),
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.String(resource["name"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Token:          gofastly.String(resource["token"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.String(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	return opts
}

func (h *HoneycombServiceAttributeHandler) buildDelete(honeycombMap any, serviceID string, serviceVersion int) *gofastly.DeleteHoneycombInput {
	resource := honeycombMap.(map[string]any)

	return &gofastly.DeleteHoneycombInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
