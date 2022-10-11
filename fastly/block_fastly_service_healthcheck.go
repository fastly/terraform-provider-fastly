package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HealthCheckServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type HealthCheckServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceHealthCheck returns a new resource.
func NewServiceHealthCheck(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&HealthCheckServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "healthcheck",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *HealthCheckServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *HealthCheckServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"check_interval": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     5000,
					Description: "How often to run the Healthcheck in milliseconds. Default `5000`",
				},
				"expected_response": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     200,
					Description: "The status code expected from the host. Default `200`",
				},
				// NOTE: We can't use TypeList as the Fastly API orders the headers.
				//
				// The Fastly API returns the contained elements in a sorted order,
				// which might otherwise cause unexpected diffs if using TypeList as the
				// data type expects the elements to be consistently ordered. So if a
				// user was to define their headers as ["b: 2", "a: 1"], then there
				// would be a constant diff as the API returns ["a: 1", "b: 2"].
				//
				// Reference:
				// https://www.terraform.io/plugin/sdkv2/schemas/schema-types#typelist
				"headers": {
					Type: schema.TypeSet,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional:    true,
					Description: "Custom health check HTTP headers (e.g. if your health check requires an API key to be provided)",
				},
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The Host header to send for this Healthcheck",
				},
				"http_version": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "1.1",
					Description: "Whether to use version 1.0 or 1.1 HTTP. Default `1.1`",
				},
				"initial": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     3,
					Description: "When loading a config, the initial number of probes to be seen as OK. Default `3`",
				},
				"method": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "HEAD",
					Description: "Which HTTP method to use. Default `HEAD`",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name to identify this Healthcheck. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"path": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The path to check",
				},
				"threshold": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     3,
					Description: "How many Healthchecks must succeed to be considered healthy. Default `3`",
				},
				"timeout": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     500,
					Description: "Timeout in milliseconds. Default `500`",
				},
				"window": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     5,
					Description: "The number of most recent Healthcheck queries to keep for this Healthcheck. Default `5`",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *HealthCheckServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	hi := resource["headers"].(*schema.Set).List()
	var hs []string
	for _, v := range hi {
		hs = append(hs, v.(string))
	}

	opts := gofastly.CreateHealthCheckInput{
		ServiceID:        d.Id(),
		ServiceVersion:   serviceVersion,
		Name:             resource["name"].(string),
		Headers:          hs,
		Host:             resource["host"].(string),
		Path:             resource["path"].(string),
		CheckInterval:    gofastly.Uint(uint(resource["check_interval"].(int))),
		ExpectedResponse: gofastly.Uint(uint(resource["expected_response"].(int))),
		HTTPVersion:      resource["http_version"].(string),
		Initial:          gofastly.Uint(uint(resource["initial"].(int))),
		Method:           resource["method"].(string),
		Threshold:        gofastly.Uint(uint(resource["threshold"].(int))),
		Timeout:          gofastly.Uint(uint(resource["timeout"].(int))),
		Window:           gofastly.Uint(uint(resource["window"].(int))),
	}

	log.Printf("[DEBUG] Create Healthcheck Opts: %#v", opts)
	_, err := conn.CreateHealthCheck(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *HealthCheckServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Healthcheck for (%s)", d.Id())
		healthcheckList, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Healthcheck for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		hcl := flattenHealthchecks(healthcheckList)

		if err := d.Set(h.GetKey(), hcl); err != nil {
			log.Printf("[WARN] Error setting Healthcheck for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *HealthCheckServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateHealthCheckInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between any we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["comment"]; ok {
		opts.Comment = gofastly.String(v.(string))
	}
	if v, ok := modified["method"]; ok {
		opts.Method = gofastly.String(v.(string))
	}
	if v, ok := modified["host"]; ok {
		opts.Host = gofastly.String(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.String(v.(string))
	}
	if v, ok := modified["http_version"]; ok {
		opts.HTTPVersion = gofastly.String(v.(string))
	}
	if v, ok := modified["timeout"]; ok {
		opts.Timeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["check_interval"]; ok {
		opts.CheckInterval = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["expected_response"]; ok {
		opts.ExpectedResponse = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["window"]; ok {
		opts.Window = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["threshold"]; ok {
		opts.Threshold = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["initial"]; ok {
		opts.Initial = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["headers"]; ok {
		h, ok := v.(*schema.Set)
		if ok {
			hi := h.List()
			var hs []string
			for _, hv := range hi {
				hs = append(hs, hv.(string))
			}
			opts.Headers = &hs
		}
	}

	log.Printf("[DEBUG] Update Healthcheck Opts: %#v", opts)
	_, err := conn.UpdateHealthCheck(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HealthCheckServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteHealthCheckInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Healthcheck removal opts: %#v", opts)
	err := conn.DeleteHealthCheck(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func flattenHealthchecks(healthcheckList []*gofastly.HealthCheck) []map[string]any {
	var hl []map[string]any
	for _, h := range healthcheckList {
		// Convert HealthChecks to a map for saving to state.
		nh := map[string]any{
			"name":              h.Name,
			"headers":           h.Headers,
			"host":              h.Host,
			"path":              h.Path,
			"check_interval":    h.CheckInterval,
			"expected_response": h.ExpectedResponse,
			"http_version":      h.HTTPVersion,
			"initial":           h.Initial,
			"method":            h.Method,
			"threshold":         h.Threshold,
			"timeout":           h.Timeout,
			"window":            h.Window,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}

	return hl
}
