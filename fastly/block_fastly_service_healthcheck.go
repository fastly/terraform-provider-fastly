package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
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
					Description: "Custom health check HTTP headers (e.g. if your health check requires an API key to be provided).",
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
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Headers:          &hs,
		Host:             gofastly.ToPointer(resource["host"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		CheckInterval:    gofastly.ToPointer(resource["check_interval"].(int)),
		ExpectedResponse: gofastly.ToPointer(resource["expected_response"].(int)),
		HTTPVersion:      gofastly.ToPointer(resource["http_version"].(string)),
		Initial:          gofastly.ToPointer(resource["initial"].(int)),
		Method:           gofastly.ToPointer(resource["method"].(string)),
		Threshold:        gofastly.ToPointer(resource["threshold"].(int)),
		Timeout:          gofastly.ToPointer(resource["timeout"].(int)),
		Window:           gofastly.ToPointer(resource["window"].(int)),
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
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Healthcheck for (%s)", d.Id())
		remoteState, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Healthcheck for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		hcl := flattenHealthchecks(remoteState)

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

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["comment"]; ok {
		opts.Comment = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["method"]; ok {
		opts.Method = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["host"]; ok {
		opts.Host = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["http_version"]; ok {
		opts.HTTPVersion = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["timeout"]; ok {
		opts.Timeout = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["check_interval"]; ok {
		opts.CheckInterval = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["expected_response"]; ok {
		opts.ExpectedResponse = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["window"]; ok {
		opts.Window = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["threshold"]; ok {
		opts.Threshold = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["initial"]; ok {
		opts.Initial = gofastly.ToPointer(v.(int))
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

// flattenHealthchecks models data into format suitable for saving to Terraform state.
func flattenHealthchecks(remoteState []*gofastly.HealthCheck) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Headers != nil {
			data["headers"] = resource.Headers
		}
		if resource.Host != nil {
			data["host"] = *resource.Host
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.CheckInterval != nil {
			data["check_interval"] = *resource.CheckInterval
		}
		if resource.ExpectedResponse != nil {
			data["expected_response"] = *resource.ExpectedResponse
		}
		if resource.HTTPVersion != nil {
			data["http_version"] = *resource.HTTPVersion
		}
		if resource.Initial != nil {
			data["initial"] = *resource.Initial
		}
		if resource.Method != nil {
			data["method"] = *resource.Method
		}
		if resource.Threshold != nil {
			data["threshold"] = *resource.Threshold
		}
		if resource.Timeout != nil {
			data["timeout"] = *resource.Timeout
		}
		if resource.Window != nil {
			data["window"] = *resource.Window
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
