package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type HoneycombServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingHoneycomb() ServiceAttributeDefinition {
	return &HoneycombServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "logging_honeycomb",
		},
	}
}

func (h *HoneycombServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client, serviceType string) error {
	serviceID := d.Id()
	ol, nl := d.GetChange(h.GetKey())

	if ol == nil {
		ol = new(schema.Set)
	}
	if nl == nil {
		nl = new(schema.Set)
	}

	ols := ol.(*schema.Set)
	nls := nl.(*schema.Set)

	removeHoneycombLogging := ols.Difference(nls).List()
	addHoneycombLogging := nls.Difference(ols).List()

	// DELETE old Honeycomb logging endpoints.
	for _, oRaw := range removeHoneycombLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteHoneycomb(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Honeycomb logging endpoint removal opts: %#v", opts)

		if err := deleteHoneycomb(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Honeycomb logging endpoints.
	for _, nRaw := range addHoneycombLogging {
		lf := nRaw.(map[string]interface{})
		opts := buildCreateHoneycomb(lf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Honeycomb logging addition opts: %#v", opts)

		if err := createHoneycomb(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *HoneycombServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client, serviceType string) error {
	// Refresh Honeycomb.
	log.Printf("[DEBUG] Refreshing Honeycomb logging endpoints for (%s)", d.Id())
	honeycombList, err := conn.ListHoneycombs(&gofastly.ListHoneycombsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Honeycomb logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenHoneycomb(honeycombList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting Honeycomb logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createHoneycomb(conn *gofastly.Client, i *gofastly.CreateHoneycombInput) error {
	_, err := conn.CreateHoneycomb(i)
	return err
}

func deleteHoneycomb(conn *gofastly.Client, i *gofastly.DeleteHoneycombInput) error {
	err := conn.DeleteHoneycomb(i)

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

func flattenHoneycomb(honeycombList []*gofastly.Honeycomb) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range honeycombList {
		// Convert Honeycomb logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"token":              ll.Token,
			"dataset":            ll.Dataset,
			"format":             ll.Format,
			"format_version":     ll.FormatVersion,
			"placement":          ll.Placement,
			"response_condition": ll.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range nll {
			if v == "" {
				delete(nll, k)
			}
		}

		lsl = append(lsl, nll)
	}

	return lsl
}

func buildCreateHoneycomb(honeycombMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateHoneycombInput {
	df := honeycombMap.(map[string]interface{})

	return &gofastly.CreateHoneycombInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		Token:             gofastly.NullString(df["token"].(string)),
		Dataset:           gofastly.NullString(df["dataset"].(string)),
		Format:            gofastly.NullString(df["format"].(string)),
		FormatVersion:     gofastly.Uint(uint(df["format_version"].(int))),
		Placement:         gofastly.NullString(df["placement"].(string)),
		ResponseCondition: gofastly.NullString(df["response_condition"].(string)),
	}
}

func buildDeleteHoneycomb(honeycombMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteHoneycombInput {
	df := honeycombMap.(map[string]interface{})

	return &gofastly.DeleteHoneycombInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}

func (h *HoneycombServiceAttributeHandler) Register(s *schema.Resource, serviceType string) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The unique name of the Honeycomb logging endpoint.",
				},

				"token": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The Write Key from the Account page of your Honeycomb account.",
				},

				"dataset": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The Honeycomb Dataset you want to log to.",
				},

				// Optional fields
				"format": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Apache style log formatting. Your log must produce valid JSON that Honeycomb can ingest.",
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
					Description:  "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
					ValidateFunc: validateLoggingPlacement(),
				},

				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
				},
			},
		},
	}
	return nil
}
