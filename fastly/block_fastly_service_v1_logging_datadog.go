package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DatadogServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingDatadog() ServiceAttributeDefinition {
	return &DatadogServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "logging_datadog",
		},
	}
}

func (h *DatadogServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	od, nd := d.GetChange(h.GetKey())

	if od == nil {
		od = new(schema.Set)
	}
	if nd == nil {
		nd = new(schema.Set)
	}

	ods := od.(*schema.Set)
	nds := nd.(*schema.Set)

	removeDatadogLogging := ods.Difference(nds).List()
	addDatadogLogging := nds.Difference(ods).List()

	// DELETE old Datadog logging endpoints.
	for _, oRaw := range removeDatadogLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteDatadog(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Datadog logging endpoint removal opts: %#v", opts)

		if err := deleteDatadog(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Datadog logging endpoints.
	for _, nRaw := range addDatadogLogging {
		df := nRaw.(map[string]interface{})
		opts := buildCreateDatadog(df, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Datadog logging addition opts: %#v", opts)

		if err := createDatadog(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *DatadogServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh Datadog.
	log.Printf("[DEBUG] Refreshing Datadog logging endpoints for (%s)", d.Id())
	datadogList, err := conn.ListDatadog(&gofastly.ListDatadogInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Datadog logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	dll := flattenDatadog(datadogList)

	if err := d.Set(h.GetKey(), dll); err != nil {
		log.Printf("[WARN] Error setting Datadog logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createDatadog(conn *gofastly.Client, i *gofastly.CreateDatadogInput) error {
	_, err := conn.CreateDatadog(i)
	return err
}

func deleteDatadog(conn *gofastly.Client, i *gofastly.DeleteDatadogInput) error {
	err := conn.DeleteDatadog(i)

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

func (h *DatadogServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The unique name of the Datadog logging endpoint.",
				},

				"token": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The API key from your Datadog account.",
				},

				// Optional fields
				"region": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "US",
					Description: "The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined.",
				},

				"format": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Apache-style string or VCL variables to use for log formatting.",
				},

				"format_version": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      2,
					Description:  "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
					ValidateFunc: validateLoggingFormatVersion(),
				},

				"placement": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Where in the generated VCL the logging call should be placed.",
					ValidateFunc: validateLoggingPlacement(),
				},

				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name of the condition to apply.",
				},
			},
		},
	}
	return nil
}

func flattenDatadog(datadogList []*gofastly.Datadog) []map[string]interface{} {
	var dsl []map[string]interface{}
	for _, dl := range datadogList {
		// Convert Datadog logging to a map for saving to state.
		ndl := map[string]interface{}{
			"name":               dl.Name,
			"token":              dl.Token,
			"region":             dl.Region,
			"format":             dl.Format,
			"format_version":     dl.FormatVersion,
			"placement":          dl.Placement,
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

func buildCreateDatadog(datadogMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateDatadogInput {
	df := datadogMap.(map[string]interface{})

	return &gofastly.CreateDatadogInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		Token:             gofastly.NullString(df["token"].(string)),
		Region:            gofastly.NullString(df["region"].(string)),
		Format:            gofastly.NullString(df["format"].(string)),
		FormatVersion:     gofastly.Uint(uint(df["format_version"].(int))),
		Placement:         gofastly.NullString(df["placement"].(string)),
		ResponseCondition: gofastly.NullString(df["response_condition"].(string)),
	}
}

func buildDeleteDatadog(datadogMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteDatadogInput {
	df := datadogMap.(map[string]interface{})

	return &gofastly.DeleteDatadogInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
