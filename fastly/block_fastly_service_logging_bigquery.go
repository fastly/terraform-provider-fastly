package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// BigQueryLoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type BigQueryLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingBigQuery returns a new resource.
func NewServiceLoggingBigQuery(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&BigQueryLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_bigquery",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *BigQueryLoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *BigQueryLoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"dataset": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your BigQuery dataset",
		},
		"email": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_BQ_EMAIL", ""),
			Description: "The email for the service account with write access to your BigQuery dataset. If not provided, this will be pulled from a `FASTLY_BQ_EMAIL` environment variable",
			Sensitive:   true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify this BigQuery logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"project_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your GCP project",
		},
		"secret_key": {
			Type:             schema.TypeString,
			Required:         true,
			DefaultFunc:      schema.EnvDefaultFunc("FASTLY_BQ_SECRET_KEY", ""),
			Description:      "The secret key associated with the service account that has write access to your BigQuery table. If not provided, this will be pulled from the `FASTLY_BQ_SECRET_KEY` environment variable. Typical format for this is a private key in a string with newlines",
			Sensitive:        true,
			ValidateDiagFunc: validateStringTrimmed,
		},
		"table": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your BigQuery table",
		},
		"template": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "BigQuery table name suffix template",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The logging format desired.",
			Default:     "%h %l %u %t \"%r\" %>s %b",
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition to apply this logging.",
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
func (h *BigQueryLoggingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateBigQueryInput{
		ServiceID:         d.Id(),
		ServiceVersion:    serviceVersion,
		Name:              gofastly.String(resource["name"].(string)),
		ProjectID:         gofastly.String(resource["project_id"].(string)),
		Dataset:           gofastly.String(resource["dataset"].(string)),
		Table:             gofastly.String(resource["table"].(string)),
		User:              gofastly.String(resource["email"].(string)),
		SecretKey:         gofastly.String(resource["secret_key"].(string)),
		Template:          gofastly.String(resource["template"].(string)),
		ResponseCondition: gofastly.String(vla.responseCondition),
		Placement:         gofastly.String(vla.placement),
	}

	if vla.format != "" {
		opts.Format = gofastly.String(vla.format)
	}

	log.Printf("[DEBUG] Create BigQuery opts: %#v", opts)
	_, err := conn.CreateBigQuery(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *BigQueryLoggingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing BigQuery for (%s)", d.Id())
		bqs, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up BigQuery logging for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		bql := flattenBigQuery(bqs)

		for _, element := range bql {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), bql); err != nil {
			log.Printf("[WARN] Error setting BigQuery for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *BigQueryLoggingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateBigQueryInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	if v, ok := modified["project_id"]; ok {
		opts.ProjectID = gofastly.String(v.(string))
	}
	if v, ok := modified["dataset"]; ok {
		opts.Dataset = gofastly.String(v.(string))
	}
	if v, ok := modified["table"]; ok {
		opts.Table = gofastly.String(v.(string))
	}
	if v, ok := modified["template_suffix"]; ok {
		opts.Template = gofastly.String(v.(string))
	}
	if v, ok := modified["email"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}

	log.Printf("[DEBUG] Update BigQuery Opts: %#v", opts)
	_, err := conn.UpdateBigQuery(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the resource.
func (h *BigQueryLoggingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteBigQueryInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly BigQuery removal opts: %#v", opts)
	err := conn.DeleteBigQuery(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func flattenBigQuery(bqList []*gofastly.BigQuery) []map[string]any {
	var sm []map[string]any
	for _, currentBQ := range bqList {
		// Convert gcs to a map for saving to state.
		m := map[string]any{
			"name":               currentBQ.Name,
			"format":             currentBQ.Format,
			"email":              currentBQ.User,
			"secret_key":         currentBQ.SecretKey,
			"project_id":         currentBQ.ProjectID,
			"dataset":            currentBQ.Dataset,
			"table":              currentBQ.Table,
			"response_condition": currentBQ.ResponseCondition,
			"template":           currentBQ.Template,
			"placement":          currentBQ.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range m {
			if v == "" {
				delete(m, k)
			}
		}

		sm = append(sm, m)
	}

	return sm
}
