package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewRelicOTLPServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type NewRelicOTLPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingNewRelicOTLP returns a new resource.
func NewServiceLoggingNewRelicOTLP(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&NewRelicOTLPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_newrelicotlp",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *NewRelicOTLPServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *NewRelicOTLPServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting. Your log must produce valid JSON that New Relic OTLP can ingest.",
		},
		"format_version": {
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the New Relic OTLP logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"placement": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. Default: `US`",
		},
		"response_condition": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The Insert API key from the Account page of your New Relic account",
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The optional New Relic Trace Observer URL to stream logs to for Infinite Tracing.",
		},
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
func (h *NewRelicOTLPServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly New Relic OTLP logging addition opts: %#v", opts)

	_, err := conn.CreateNewRelicOTLP(opts)
	return err
}

// Read refreshes the resource.
func (h *NewRelicOTLPServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing New Relic OTLP logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListNewRelicOTLP(&gofastly.ListNewRelicOTLPInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up New Relic OTLP logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dll := flattenNewRelicOTLP(remoteState)

		if err := d.Set(h.GetKey(), dll); err != nil {
			log.Printf("[WARN] Error setting New Relic OTLP logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *NewRelicOTLPServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateNewRelicOTLPInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update New Relic OTLP Opts: %#v", opts)
	_, err := conn.UpdateNewRelicOTLP(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *NewRelicOTLPServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := &gofastly.DeleteNewRelicOTLPInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly New Relic OTLP logging endpoint removal opts: %#v", opts)

	err := conn.DeleteNewRelicOTLP(opts)

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

// flattenNewRelicOTLP models data into format suitable for saving to Terraform state.
func flattenNewRelicOTLP(remoteState []*gofastly.NewRelicOTLP) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"name":               resource.Name,
			"placement":          resource.Placement,
			"region":             resource.Region,
			"response_condition": resource.ResponseCondition,
			"token":              resource.Token,
			"url":                resource.URL,
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

func (h *NewRelicOTLPServiceAttributeHandler) buildCreate(newrelicotlpMap any, serviceID string, serviceVersion int) *gofastly.CreateNewRelicOTLPInput {
	resource := newrelicotlpMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateNewRelicOTLPInput{
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Region:         gofastly.ToPointer(resource["region"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Token:          gofastly.ToPointer(resource["token"].(string)),
		URL:            gofastly.ToPointer(resource["url"].(string)),
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
