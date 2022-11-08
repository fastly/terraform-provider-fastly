package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// LogglyServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type LogglyServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingLoggly returns a new resource.
func NewServiceLoggingLoggly(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&LogglyServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_loggly",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *LogglyServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *LogglyServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Loggly logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},

		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/).",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
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
func (h *LogglyServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Loggly logging addition opts: %#v", opts)

	return createLoggly(conn, opts)
}

// Read refreshes the resource.
func (h *LogglyServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Loggly logging endpoints for (%s)", d.Id())
		logglyList, err := conn.ListLoggly(&gofastly.ListLogglyInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Loggly logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenLoggly(logglyList)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting Loggly logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *LogglyServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateLogglyInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Loggly Opts: %#v", opts)
	_, err := conn.UpdateLoggly(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *LogglyServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Loggly logging endpoint removal opts: %#v", opts)

	return deleteLoggly(conn, opts)
}

func createLoggly(conn *gofastly.Client, i *gofastly.CreateLogglyInput) error {
	_, err := conn.CreateLoggly(i)
	return err
}

func deleteLoggly(conn *gofastly.Client, i *gofastly.DeleteLogglyInput) error {
	err := conn.DeleteLoggly(i)

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

func flattenLoggly(logglyList []*gofastly.Loggly) []map[string]any {
	var lsl []map[string]any
	for _, ll := range logglyList {
		// Convert Loggly logging to a map for saving to state.
		nll := map[string]any{
			"name":               ll.Name,
			"token":              ll.Token,
			"format":             ll.Format,
			"format_version":     ll.FormatVersion,
			"placement":          ll.Placement,
			"response_condition": ll.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range nll {
			if v == "" {
				delete(nll, k)
			}
		}

		lsl = append(lsl, nll)
	}

	return lsl
}

func (h *LogglyServiceAttributeHandler) buildCreate(logglyMap any, serviceID string, serviceVersion int) *gofastly.CreateLogglyInput {
	df := logglyMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateLogglyInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              gofastly.String(df["name"].(string)),
		Token:             gofastly.String(df["token"].(string)),
		Format:            gofastly.String(vla.format),
		FormatVersion:     gofastly.Int(intOrDefault(vla.formatVersion)),
		Placement:         gofastly.String(vla.placement),
		ResponseCondition: gofastly.String(vla.responseCondition),
	}
}

func (h *LogglyServiceAttributeHandler) buildDelete(logglyMap any, serviceID string, serviceVersion int) *gofastly.DeleteLogglyInput {
	df := logglyMap.(map[string]any)

	return &gofastly.DeleteLogglyInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
