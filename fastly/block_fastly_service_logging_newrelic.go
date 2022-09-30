package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewRelicServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type NewRelicServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingNewRelic returns a new resource.
func NewServiceLoggingNewRelic(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&NewRelicServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_newrelic",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *NewRelicServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *NewRelicServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the New Relic logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The Insert API key from the Account page of your New Relic account",
		},
		// Optional
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. Default: `US`",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting. Your log must produce valid JSON that New Relic Logs can ingest.",
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
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
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
func (h *NewRelicServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly New Relic logging addition opts: %#v", opts)

	return createNewRelic(conn, opts)
}

// Read refreshes the resource.
func (h *NewRelicServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing New Relic logging endpoints for (%s)", d.Id())
		newrelicList, err := conn.ListNewRelic(&gofastly.ListNewRelicInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up New Relic logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dll := flattenNewRelic(newrelicList)

		for _, element := range dll {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), dll); err != nil {
			log.Printf("[WARN] Error setting New Relic logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *NewRelicServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateNewRelicInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between interface{} we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update New Relic Opts: %#v", opts)
	_, err := conn.UpdateNewRelic(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *NewRelicServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly New Relic logging endpoint removal opts: %#v", opts)

	return deleteNewRelic(conn, opts)
}

func createNewRelic(conn *gofastly.Client, i *gofastly.CreateNewRelicInput) error {
	_, err := conn.CreateNewRelic(i)
	return err
}

func deleteNewRelic(conn *gofastly.Client, i *gofastly.DeleteNewRelicInput) error {
	err := conn.DeleteNewRelic(i)

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

func flattenNewRelic(newrelicList []*gofastly.NewRelic) []map[string]interface{} {
	var dsl []map[string]interface{}
	for _, dl := range newrelicList {
		// Convert NewRelic logging to a map for saving to state.
		ndl := map[string]interface{}{
			"name":               dl.Name,
			"token":              dl.Token,
			"format":             dl.Format,
			"format_version":     dl.FormatVersion,
			"placement":          dl.Placement,
			"region":             dl.Region,
			"response_condition": dl.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range ndl {
			if v == "" {
				delete(ndl, k)
			}
		}

		dsl = append(dsl, ndl)
	}

	return dsl
}

func (h *NewRelicServiceAttributeHandler) buildCreate(newrelicMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateNewRelicInput {
	df := newrelicMap.(map[string]interface{})

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateNewRelicInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		Token:             df["token"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		Region:            df["region"].(string),
		ResponseCondition: vla.responseCondition,
	}
}

func (h *NewRelicServiceAttributeHandler) buildDelete(newrelicMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteNewRelicInput {
	df := newrelicMap.(map[string]interface{})

	return &gofastly.DeleteNewRelicInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
