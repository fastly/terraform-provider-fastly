package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SumologicServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SumologicServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingSumologic returns a new resource.
func NewServiceLoggingSumologic(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&SumologicServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_sumologic",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *SumologicServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *SumologicServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"message_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "classic",
			Description:      MessageTypeDescription,
			ValidateDiagFunc: validateLoggingMessageType(),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify this Sumologic endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The URL to Sumologic collector endpoint",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     `%h %l %u %t "%r" %>s %b`,
			Description: "Apache-style string or VCL variables to use for log formatting",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 2)",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of blockAttributes condition to apply this logging.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
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
func (h *SumologicServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateSumologicInput{
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		MessageType:    gofastly.String(resource["message_type"].(string)),
		Name:           gofastly.String(resource["name"].(string)),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		URL:            gofastly.String(resource["url"].(string)),
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

	log.Printf("[DEBUG] Create Sumologic Opts: %#v", opts)
	_, err := conn.CreateSumologic(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *SumologicServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Sumologic for (%s)", d.Id())
		remoteState, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Sumologic for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		sul := flattenSumologics(remoteState)

		for _, element := range sul {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), sul); err != nil {
			log.Printf("[WARN] Error setting Sumologic for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *SumologicServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSumologicInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.String(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Sumologic Opts: %#v", opts)
	_, err := conn.UpdateSumologic(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *SumologicServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSumologicInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Sumologic removal opts: %#v", opts)
	err := conn.DeleteSumologic(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenSumologics models data into format suitable for saving to Terraform state.
func flattenSumologics(remoteState []*gofastly.Sumologic) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"url":                resource.URL,
			"format":             resource.Format,
			"response_condition": resource.ResponseCondition,
			"message_type":       resource.MessageType,
			"format_version":     int(resource.FormatVersion),
			"placement":          resource.Placement,
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
