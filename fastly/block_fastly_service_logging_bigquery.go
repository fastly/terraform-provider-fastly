package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
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
		"account_name": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_ACCOUNT_NAME", ""),
			Description: "The google account name used to obtain temporary credentials (default none). You may optionally provide this via an environment variable, `FASTLY_GCS_ACCOUNT_NAME`.",
		},
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
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Where in the generated VCL the logging call should be placed (ignored).",
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
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		ProjectID:      gofastly.ToPointer(resource["project_id"].(string)),
		Dataset:        gofastly.ToPointer(resource["dataset"].(string)),
		Table:          gofastly.ToPointer(resource["table"].(string)),
		User:           gofastly.ToPointer(resource["email"].(string)),
		SecretKey:      gofastly.ToPointer(resource["secret_key"].(string)),
		Template:       gofastly.ToPointer(resource["template"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.format != "" {
		opts.Format = gofastly.ToPointer(vla.format)
	}
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
	}
	if v, ok := resource["account_name"].(string); ok && v != "" {
		opts.AccountName = gofastly.ToPointer(v)
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
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing BigQuery for (%s)", d.Id())
		remoteState, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up BigQuery logging for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		bql := flattenBigQuery(remoteState)

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

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["project_id"]; ok {
		opts.ProjectID = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["dataset"]; ok {
		opts.Dataset = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["table"]; ok {
		opts.Table = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["template_suffix"]; ok {
		opts.Template = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["email"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.ToPointer(v.(string))
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
	if v, ok := modified["account_name"]; ok {
		opts.AccountName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
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

// flattenBigQuery models data into format suitable for saving to Terraform state.
func flattenBigQuery(remoteState []*gofastly.BigQuery) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.User != nil {
			data["email"] = *resource.User
		}
		if resource.SecretKey != nil {
			data["secret_key"] = *resource.SecretKey
		}
		if resource.ProjectID != nil {
			data["project_id"] = *resource.ProjectID
		}
		if resource.Dataset != nil {
			data["dataset"] = *resource.Dataset
		}
		if resource.Table != nil {
			data["table"] = *resource.Table
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.Template != nil {
			data["template"] = *resource.Template
		}
		if resource.AccountName != nil {
			data["account_name"] = *resource.AccountName
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
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
