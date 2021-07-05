package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type BigQueryLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceBigQueryLogging(sa ServiceMetadata) ServiceAttributeDefinition {
	return &BigQueryLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "bigquerylogging",
			serviceMetadata: sa,
		},
	}
}

func (h *BigQueryLoggingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oldSet := os.(*schema.Set)
	newSet := ns.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		t, ok := resource.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
		}
		return t["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// DELETE removed resources
	for _, resource := range diffResult.Deleted {
		resource := resource.(map[string]interface{})
		opts := gofastly.DeleteBigQueryInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		log.Printf("[DEBUG] Fastly bigquerylogging removal opts: %#v", opts)
		err := conn.DeleteBigQuery(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// CREATE new resources
	for _, resource := range diffResult.Added {
		resource := resource.(map[string]interface{})
		var vla = h.getVCLLoggingAttributes(resource)
		opts := gofastly.CreateBigQueryInput{
			ServiceID:         d.Id(),
			ServiceVersion:    latestVersion,
			Name:              resource["name"].(string),
			ProjectID:         resource["project_id"].(string),
			Dataset:           resource["dataset"].(string),
			Table:             resource["table"].(string),
			User:              resource["email"].(string),
			SecretKey:         resource["secret_key"].(string),
			Template:          resource["template"].(string),
			ResponseCondition: vla.responseCondition,
			Placement:         vla.placement,
		}

		if vla.format != "" {
			opts.Format = vla.format
		}

		log.Printf("[DEBUG] Create bigquerylogging opts: %#v", opts)
		_, err := conn.CreateBigQuery(&opts)
		if err != nil {
			return err
		}
	}

	// UPDATE modified resources
	//
	// NOTE: although the go-fastly API client enables updating of a resource by
	// its 'name' attribute, this isn't possible within terraform due to
	// constraints in the data model/schema of the resources not having a uid.
	for _, resource := range diffResult.Modified {
		resource := resource.(map[string]interface{})

		opts := gofastly.UpdateBigQueryInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		// only attempt to update attributes that have changed
		modified := setDiff.Filter(resource, oldSet)

		// NOTE: where we transition between interface{} we lose the ability to
		// infer the underlying type being either a uint vs an int. This
		// materializes as a panic (yay) and so it's only at runtime we discover
		// this and so we've updated the below code to convert the type asserted
		// int into a uint before passing the value to gofastly.Uint().
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
		if v, ok := modified["user"]; ok {
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
		if v, ok := modified["format_version"]; ok {
			opts.FormatVersion = gofastly.Uint(uint(v.(int)))
		}

		log.Printf("[DEBUG] Update BigQuery Opts: %#v", opts)
		_, err := conn.UpdateBigQuery(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *BigQueryLoggingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing BigQuery for (%s)", d.Id())
	BQList, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up BigQuery logging for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	bql := flattenBigQuery(BQList)

	for _, element := range bql {
		element = h.pruneVCLLoggingAttributes(element)
	}

	if err := d.Set(h.GetKey(), bql); err != nil {
		log.Printf("[WARN] Error setting bigquerylogging for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *BigQueryLoggingServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
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
		"dataset": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your BigQuery dataset",
		},
		"table": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your BigQuery table",
		},
		"email": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_BQ_EMAIL", ""),
			Description: "The email for the service account with write access to your BigQuery dataset. If not provided, this will be pulled from a `FASTLY_BQ_EMAIL` environment variable",
			Sensitive:   true,
		},
		"secret_key": {
			Type:             schema.TypeString,
			Required:         true,
			DefaultFunc:      schema.EnvDefaultFunc("FASTLY_BQ_SECRET_KEY", ""),
			Description:      "The secret key associated with the service account that has write access to your BigQuery table. If not provided, this will be pulled from the `FASTLY_BQ_SECRET_KEY` environment variable. Typical format for this is a private key in a string with newlines",
			Sensitive:        true,
			ValidateDiagFunc: validateStringTrimmed(),
		},
		// Optional fields
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

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
	return nil
}

func flattenBigQuery(bqList []*gofastly.BigQuery) []map[string]interface{} {
	var BQList []map[string]interface{}
	for _, currentBQ := range bqList {
		// Convert gcs to a map for saving to state.
		BQMapString := map[string]interface{}{
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
		for k, v := range BQMapString {
			if v == "" {
				delete(BQMapString, k)
			}
		}

		BQList = append(BQList, BQMapString)
	}

	return BQList
}
