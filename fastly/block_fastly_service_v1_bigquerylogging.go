package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeBigquerylogging := oss.Difference(nss).List()
	addBigquerylogging := nss.Difference(oss).List()

	// DELETE old bigquerylogging configurations
	for _, pRaw := range removeBigquerylogging {
		sf := pRaw.(map[string]interface{})
		opts := gofastly.DeleteBigQueryInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           sf["name"].(string),
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

	// POST new/updated bigquerylogging
	for _, pRaw := range addBigquerylogging {
		sf := pRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("bigquerylogging")` in this case.
		if v, ok := sf["name"]; !ok || v.(string) == "" {
			continue
		}

		var vla = h.getVCLLoggingAttributes(sf)
		opts := gofastly.CreateBigQueryInput{
			ServiceID:         d.Id(),
			ServiceVersion:    latestVersion,
			Name:              sf["name"].(string),
			ProjectID:         sf["project_id"].(string),
			Dataset:           sf["dataset"].(string),
			Table:             sf["table"].(string),
			User:              sf["email"].(string),
			SecretKey:         sf["secret_key"].(string),
			Template:          sf["template"].(string),
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
			Description: "A unique name to identify this BigQuery logging endpoint",
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
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_BQ_SECRET_KEY", ""),
			Description: "The secret key associated with the service account that has write access to your BigQuery table. If not provided, this will be pulled from the `FASTLY_BQ_SECRET_KEY` environment variable. Typical format for this is a private key in a string with newlines",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
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
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed.",
			ValidateFunc: validateLoggingPlacement(),
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
