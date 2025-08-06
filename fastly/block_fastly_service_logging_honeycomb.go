package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
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
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   !DisplaySensitiveFields,
			Description: "The Write Key from the Account page of your Honeycomb account",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     LoggingHoneycombDefaultFormat,
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
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `none`.",
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
func (h *HoneycombServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Honeycomb logging addition opts: %#v", opts)

	_, err := conn.CreateHoneycomb(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *HoneycombServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Honeycomb logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListHoneycombs(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListHoneycombsInput{
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
func (h *HoneycombServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateHoneycombInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["dataset"]; ok {
		opts.Dataset = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Honeycomb Opts: %#v", opts)
	_, err := conn.UpdateHoneycomb(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HoneycombServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Honeycomb logging endpoint removal opts: %#v", opts)

	err := conn.DeleteHoneycomb(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

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
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Token != nil {
			data["token"] = *resource.Token
		}
		if resource.Dataset != nil {
			data["dataset"] = *resource.Dataset
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.ProcessingRegion != nil {
			data["processing_region"] = *resource.ProcessingRegion
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
		Dataset:          gofastly.ToPointer(resource["dataset"].(string)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		Name:             gofastly.ToPointer(resource["name"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		Token:            gofastly.ToPointer(resource["token"].(string)),
		ProcessingRegion: gofastly.ToPointer(resource["processing_region"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
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
