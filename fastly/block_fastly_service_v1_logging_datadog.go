package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DatadogServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingDatadog(sa ServiceMetadata) ServiceAttributeDefinition {
	return &DatadogServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_datadog",
			serviceMetadata: sa,
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
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Datadog logging endpoint removal opts: %#v", opts)

		if err := deleteDatadog(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Datadog logging endpoints.
	for _, nRaw := range addDatadogLogging {
		df := nRaw.(map[string]interface{})
		opts := h.buildCreate(df, serviceID, latestVersion)

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
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
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
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Datadog logging endpoint",
		},

		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The API key from your Datadog account",
		},

		// Optional fields
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
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

func (h *DatadogServiceAttributeHandler) buildCreate(datadogMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateDatadogInput {
	df := datadogMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateDatadogInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		Token:             df["token"].(string),
		Region:            df["region"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *DatadogServiceAttributeHandler) buildDelete(datadogMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteDatadogInput {
	df := datadogMap.(map[string]interface{})

	return &gofastly.DeleteDatadogInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
