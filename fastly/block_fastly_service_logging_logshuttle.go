package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// LogshuttleServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type LogshuttleServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingLogshuttle returns a new resource.
func NewServiceLoggingLogshuttle(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&LogshuttleServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_logshuttle",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *LogshuttleServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *LogshuttleServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Log Shuttle logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true, // FIXME: This isn't required in go-fastly.
			Sensitive:   true,
			Description: "The data authentication token associated with this endpoint",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your Log Shuttle endpoint URL",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
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
func (h *LogshuttleServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Log Shuttle logging addition opts: %#v", opts)

	return createLogshuttle(conn, opts)
}

// Read refreshes the resource.
func (h *LogshuttleServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Log Shuttle logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListLogshuttles(&gofastly.ListLogshuttlesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Log Shuttle logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenLogshuttle(remoteState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting Log Shuttle logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *LogshuttleServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateLogshuttleInput{
		Name:           resource["name"].(string),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Log Shuttle Opts: %#v", opts)
	_, err := conn.UpdateLogshuttle(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *LogshuttleServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Log Shuttle logging endpoint removal opts: %#v", opts)

	return deleteLogshuttle(conn, opts)
}

func createLogshuttle(conn *gofastly.Client, i *gofastly.CreateLogshuttleInput) error {
	_, err := conn.CreateLogshuttle(i)
	return err
}

func deleteLogshuttle(conn *gofastly.Client, i *gofastly.DeleteLogshuttleInput) error {
	err := conn.DeleteLogshuttle(i)

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

// flattenLogshuttle models data into format suitable for saving to Terraform state.
func flattenLogshuttle(remoteState []*gofastly.Logshuttle) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"token":              resource.Token,
			"url":                resource.URL,
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

func (h *LogshuttleServiceAttributeHandler) buildCreate(logshuttleMap any, serviceID string, serviceVersion int) *gofastly.CreateLogshuttleInput {
	resource := logshuttleMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateLogshuttleInput{
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.String(resource["name"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Token:          gofastly.String(resource["token"].(string)),
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

	return opts
}

func (h *LogshuttleServiceAttributeHandler) buildDelete(logshuttleMap any, serviceID string, serviceVersion int) *gofastly.DeleteLogshuttleInput {
	resource := logshuttleMap.(map[string]any)

	return &gofastly.DeleteLogshuttleInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
