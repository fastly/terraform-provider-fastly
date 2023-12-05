package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PaperTrailServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type PaperTrailServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingPaperTrail returns a new resource.
func NewServiceLoggingPaperTrail(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&PaperTrailServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_papertrail",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *PaperTrailServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *PaperTrailServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The address of the Papertrail endpoint",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify this Papertrail endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"port": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The port associated with the address where the Papertrail endpoint can be accessed",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     `%h %l %u %t "%r" %>s %b`,
			Description: "A Fastly [log format string](https://docs.fastly.com/en/guides/custom-log-formats)",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of 2 are placed in `vcl_log` and those with `format_version` of 1 are placed in `vcl_deliver`",
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
func (h *PaperTrailServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)

	opts := gofastly.CreatePapertrailInput{
		Address:        gofastly.ToPointer(resource["address"].(string)),
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Port:           gofastly.ToPointer(resource["port"].(int)),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
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

	log.Printf("[DEBUG] Create Papertrail Opts: %#v", opts)
	_, err := conn.CreatePapertrail(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *PaperTrailServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Papertrail for (%s)", d.Id())
		remoteState, err := conn.ListPapertrails(&gofastly.ListPapertrailsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Papertrail for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		pl := flattenPapertrails(remoteState)

		for _, element := range pl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), pl); err != nil {
			log.Printf("[WARN] Error setting Papertrail for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *PaperTrailServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdatePapertrailInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Papertrail Opts: %#v", opts)
	_, err := conn.UpdatePapertrail(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *PaperTrailServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeletePapertrailInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Papertrail removal opts: %#v", opts)
	err := conn.DeletePapertrail(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenPapertrails models data into format suitable for saving to Terraform state.
func flattenPapertrails(remoteState []*gofastly.Papertrail) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"address":            resource.Address,
			"port":               resource.Port,
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"response_condition": resource.ResponseCondition,
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
