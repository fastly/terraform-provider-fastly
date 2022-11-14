package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ScalyrServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ScalyrServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingScalyr returns a new resource.
func NewServiceLoggingScalyr(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ScalyrServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_scalyr",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ScalyrServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ScalyrServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Scalyr logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The token to use for authentication (https://www.scalyr.com/keys)",
			Sensitive:   true,
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
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
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
func (h *ScalyrServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Scalyr logging addition opts: %#v", opts)

	return createScalyr(conn, opts)
}

// Read refreshes the resource.
func (h *ScalyrServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Scalyr logging endpoints for (%s)", d.Id())
		scalyrList, err := conn.ListScalyrs(&gofastly.ListScalyrsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Scalyr logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		scalyrLogList := flattenScalyr(scalyrList)

		for _, element := range scalyrLogList {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), scalyrLogList); err != nil {
			log.Printf("[WARN] Error setting Scalyr logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ScalyrServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateScalyrInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Scalyr Opts: %#v", opts)
	_, err := conn.UpdateScalyr(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ScalyrServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Scalyr logging endpoint removal opts: %#v", opts)

	return deleteScalyr(conn, opts)
}

func createScalyr(conn *gofastly.Client, i *gofastly.CreateScalyrInput) error {
	_, err := conn.CreateScalyr(i)
	return err
}

func deleteScalyr(conn *gofastly.Client, i *gofastly.DeleteScalyrInput) error {
	err := conn.DeleteScalyr(i)

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

func flattenScalyr(scalyrList []*gofastly.Scalyr) []map[string]any {
	var flattened []map[string]any
	for _, s := range scalyrList {
		// Convert logging to a map for saving to state.
		flatScalyr := map[string]any{
			"name":               s.Name,
			"region":             s.Region,
			"token":              s.Token,
			"response_condition": s.ResponseCondition,
			"format":             s.Format,
			"placement":          s.Placement,
			"format_version":     s.FormatVersion,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range flatScalyr {
			if v == "" {
				delete(flatScalyr, k)
			}
		}

		flattened = append(flattened, flatScalyr)
	}

	return flattened
}

func (h *ScalyrServiceAttributeHandler) buildCreate(scalyrMap any, serviceID string, serviceVersion int) *gofastly.CreateScalyrInput {
	df := scalyrMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	opts := &gofastly.CreateScalyrInput{
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.String(df["name"].(string)),
		Placement:      gofastly.String(vla.placement),
		Region:         gofastly.String(df["region"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Token:          gofastly.String(df["token"].(string)),
	}

	// WARNING: The following fields shouldn't have an emptry string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	// if vla.placement != "" {
	// 	opts.Placement = gofastly.String(vla.placement)
	// }
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	return opts
}

func (h *ScalyrServiceAttributeHandler) buildDelete(scalyrMap any, serviceID string, serviceVersion int) *gofastly.DeleteScalyrInput {
	df := scalyrMap.(map[string]any)

	return &gofastly.DeleteScalyrInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
