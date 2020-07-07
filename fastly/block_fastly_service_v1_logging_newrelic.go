package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type NewRelicServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingNewRelic(sa ServiceMetadata) ServiceAttributeDefinition {
	return &NewRelicServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_newrelic",
			serviceMetadata: sa,
		},
	}
}

func (h *NewRelicServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
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

	removeNewRelicLogging := ods.Difference(nds).List()
	addNewRelicLogging := nds.Difference(ods).List()

	// DELETE old NewRelic logging endpoints.
	for _, oRaw := range removeNewRelicLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly New Relic logging endpoint removal opts: %#v", opts)

		if err := deleteNewRelic(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated NewRelic logging endpoints.
	for _, nRaw := range addNewRelicLogging {
		df := nRaw.(map[string]interface{})
		opts := h.buildCreate(df, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly New Relic logging addition opts: %#v", opts)

		if err := createNewRelic(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *NewRelicServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh NewRelic.
	log.Printf("[DEBUG] Refreshing New Relic logging endpoints for (%s)", d.Id())
	newrelicList, err := conn.ListNewRelic(&gofastly.ListNewRelicInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up New Relic logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	dll := flattenNewRelic(newrelicList)

	if err := d.Set(h.GetKey(), dll); err != nil {
		log.Printf("[WARN] Error setting New Relic logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
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

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateNewRelicInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		Token:             gofastly.NullString(df["token"].(string)),
		Format:            gofastly.NullString(vla.format),
		FormatVersion:     gofastly.Uint(vla.formatVersion),
		Placement:         gofastly.NullString(vla.placement),
		ResponseCondition: gofastly.NullString(vla.responseCondition),
	}
}

func (h *NewRelicServiceAttributeHandler) buildDelete(newrelicMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteNewRelicInput {
	df := newrelicMap.(map[string]interface{})

	return &gofastly.DeleteNewRelicInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}

func (h *NewRelicServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the New Relic logging endpoint",
		},

		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The Insert API key from the Account page of your New Relic account.",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting. Your log must produce valid JSON that New Relic Logs can ingest.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed.",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
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
