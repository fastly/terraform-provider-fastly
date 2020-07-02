package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type LogglyServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingLoggly() ServiceAttributeDefinition {
	return &LogglyServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "logging_loggly",
		},
	}
}

func (h *LogglyServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client, serviceType string) error {
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

	removeLogglyLogging := ols.Difference(nls).List()
	addLogglyLogging := nls.Difference(ols).List()

	// DELETE old Loggly logging endpoints.
	for _, oRaw := range removeLogglyLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteLoggly(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Loggly logging endpoint removal opts: %#v", opts)

		if err := deleteLoggly(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Loggly logging endpoints.
	for _, nRaw := range addLogglyLogging {
		lf := nRaw.(map[string]interface{})
		opts := buildCreateLoggly(lf, serviceID, latestVersion, serviceType)

		log.Printf("[DEBUG] Fastly Loggly logging addition opts: %#v", opts)

		if err := createLoggly(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *LogglyServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client, serviceType string) error {
	// Refresh Loggly.
	log.Printf("[DEBUG] Refreshing Loggly logging endpoints for (%s)", d.Id())
	logglyList, err := conn.ListLoggly(&gofastly.ListLogglyInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Loggly logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenLoggly(logglyList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting Loggly logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createLoggly(conn *gofastly.Client, i *gofastly.CreateLogglyInput) error {
	_, err := conn.CreateLoggly(i)
	return err
}

func deleteLoggly(conn *gofastly.Client, i *gofastly.DeleteLogglyInput) error {
	err := conn.DeleteLoggly(i)

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

func flattenLoggly(logglyList []*gofastly.Loggly) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range logglyList {
		// Convert Loggly logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"token":              ll.Token,
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

func buildCreateLoggly(logglyMap interface{}, serviceID string, serviceVersion int, serviceType string) *gofastly.CreateLogglyInput {
	df := logglyMap.(map[string]interface{})

	var vla = NewVCLLoggingAttributes()
	if serviceType == ServiceTypeVCL {
		vla.format = df["format"].(string)
		vla.formatVersion = uint(df["format_version"].(int))
		vla.placement = df["placement"].(string)
		vla.responseCondition = df["response_condition"].(string)
	}

	return &gofastly.CreateLogglyInput{
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

func buildDeleteLoggly(logglyMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteLogglyInput {
	df := logglyMap.(map[string]interface{})

	return &gofastly.DeleteLogglyInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}

func (h *LogglyServiceAttributeHandler) Register(s *schema.Resource, serviceType string) error {
	var a = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Loggly logging endpoint.",
		},

		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/).",
		},
	}

	if serviceType == ServiceTypeVCL {
		a["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
		}
		a["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		a["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
			ValidateFunc: validateLoggingPlacement(),
		}
		a["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: a,
		},
	}
	return nil
}
