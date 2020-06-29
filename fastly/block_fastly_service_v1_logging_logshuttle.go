package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type LogshuttleServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingLogshuttle() ServiceAttributeDefinition {
	return &LogshuttleServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "logging_logshuttle",
		},
	}
}

func (h *LogshuttleServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
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

	removeLogshuttleLogging := ols.Difference(nls).List()
	addLogshuttleLogging := nls.Difference(ols).List()

	// DELETE old Log Shuttle logging endpoints.
	for _, oRaw := range removeLogshuttleLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteLogshuttle(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Log Shuttle logging endpoint removal opts: %#v", opts)

		if err := deleteLogshuttle(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Log Shuttle logging endpoints.
	for _, nRaw := range addLogshuttleLogging {
		lf := nRaw.(map[string]interface{})
		opts := buildCreateLogshuttle(lf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Log Shuttle logging addition opts: %#v", opts)

		if err := createLogshuttle(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *LogshuttleServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh Log Shuttle.
	log.Printf("[DEBUG] Refreshing Log Shuttle logging endpoints for (%s)", d.Id())
	logshuttleList, err := conn.ListLogshuttles(&gofastly.ListLogshuttlesInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Log Shuttle logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenLogshuttle(logshuttleList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting Log Shuttle logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createLogshuttle(conn *gofastly.Client, i *gofastly.CreateLogshuttleInput) error {
	_, err := conn.CreateLogshuttle(i)
	return err
}

func deleteLogshuttle(conn *gofastly.Client, i *gofastly.DeleteLogshuttleInput) error {
	err := conn.DeleteLogshuttle(i)

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

func flattenLogshuttle(logshuttleList []*gofastly.Logshuttle) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range logshuttleList {
		// Convert Log Shuttle logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"token":              ll.Token,
			"url":                ll.URL,
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

func buildCreateLogshuttle(logshuttleMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateLogshuttleInput {
	df := logshuttleMap.(map[string]interface{})

	return &gofastly.CreateLogshuttleInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		Token:             gofastly.NullString(df["token"].(string)),
		URL:               gofastly.NullString(df["url"].(string)),
		Format:            gofastly.NullString(df["format"].(string)),
		FormatVersion:     gofastly.Uint(uint(df["format_version"].(int))),
		Placement:         gofastly.NullString(df["placement"].(string)),
		ResponseCondition: gofastly.NullString(df["response_condition"].(string)),
	}
}

func buildDeleteLogshuttle(logshuttleMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteLogshuttleInput {
	df := logshuttleMap.(map[string]interface{})

	return &gofastly.DeleteLogshuttleInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}

func (h *LogshuttleServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The unique name of the Log Shuttle logging endpoint.",
				},

				"token": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The data authentication token associated with this endpoint.",
				},

				"url": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Your Log Shuttle endpoint url.",
				},

				// Optional fields
				"format": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Apache style log formatting.",
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
