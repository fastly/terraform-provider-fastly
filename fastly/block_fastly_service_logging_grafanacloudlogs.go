package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GrafanaCloudLogsServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type GrafanaCloudLogsServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingGrafanaCloudLogs returns a new resource.
func NewServiceLoggingGrafanaCloudLogs(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&GrafanaCloudLogsServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_grafanacloudlogs",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *GrafanaCloudLogsServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *GrafanaCloudLogsServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"index": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The stream identifier as a JSON string",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the GrafanaCloudLogs logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The Access Policy Token key for your GrafanaCloudLogs account",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The URL to stream logs to",
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Grafana User ID",
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
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
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
func (h *GrafanaCloudLogsServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly GrafanaCloudLogs logging addition opts: %#v", opts)

	return createGrafanaCloudLogs(conn, opts)
}

// Read refreshes the resource.
func (h *GrafanaCloudLogsServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing GrafanaCloudLogs logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListGrafanaCloudLogs(&gofastly.ListGrafanaCloudLogsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up GrafanaCloudLogs logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dll := flattenGrafanaCloudLogs(remoteState)

		for _, element := range dll {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), dll); err != nil {
			log.Printf("[WARN] Error setting GrafanaCloudLogs logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *GrafanaCloudLogsServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateGrafanaCloudLogsInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["index"]; ok {
		opts.Index = gofastly.ToPointer(v.(string))
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

	log.Printf("[DEBUG] Update GrafanaCloudLogs Opts: %#v", opts)
	_, err := conn.UpdateGrafanaCloudLogs(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GrafanaCloudLogsServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly GrafanaCloudLogs logging endpoint removal opts: %#v", opts)

	return deleteGrafanaCloudLogs(conn, opts)
}

func createGrafanaCloudLogs(conn *gofastly.Client, i *gofastly.CreateGrafanaCloudLogsInput) error {
	_, err := conn.CreateGrafanaCloudLogs(i)
	return err
}

func deleteGrafanaCloudLogs(conn *gofastly.Client, i *gofastly.DeleteGrafanaCloudLogsInput) error {
	err := conn.DeleteGrafanaCloudLogs(i)

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

// flattenGrafanaCloudLogs models data into format suitable for saving to Terraform state.
func flattenGrafanaCloudLogs(remoteState []*gofastly.GrafanaCloudLogs) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.User != nil {
			data["user"] = *resource.User
		}
		if resource.Token != nil {
			data["token"] = *resource.Token
		}
		if resource.URL != nil {
			data["url"] = *resource.URL
		}
		if resource.Index != nil {
			data["index"] = *resource.Index
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

func (h *GrafanaCloudLogsServiceAttributeHandler) buildCreate(grafanacloudlogsMap any, serviceID string, serviceVersion int) *gofastly.CreateGrafanaCloudLogsInput {
	resource := grafanacloudlogsMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateGrafanaCloudLogsInput{
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		User:           gofastly.ToPointer(resource["user"].(string)),
		Token:          gofastly.ToPointer(resource["token"].(string)),
		URL:            gofastly.ToPointer(resource["url"].(string)),
		Index:          gofastly.ToPointer(resource["index"].(string)),
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

func (h *GrafanaCloudLogsServiceAttributeHandler) buildDelete(grafanacloudlogsMap any, serviceID string, serviceVersion int) *gofastly.DeleteGrafanaCloudLogsInput {
	resource := grafanacloudlogsMap.(map[string]any)

	return &gofastly.DeleteGrafanaCloudLogsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
