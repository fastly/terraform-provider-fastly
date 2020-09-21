package fastly

import (
	"fmt"
	"log"

	"github.com/fastly/go-fastly/fastly"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ScalyrServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingScalyr(sa ServiceMetadata) ServiceAttributeDefinition {
	return &ScalyrServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_scalyr",
			serviceMetadata: sa,
		},
	}
}

func (h *ScalyrServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Scalyr logging endpoint.",
		},

		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The token to use for authentication (https://www.scalyr.com/keys).",
			Sensitive:   true,
		},

		// Optional
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. One of US or EU. Defaults to US if undefined.",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
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
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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

func (h *ScalyrServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	oldLogCfg, newLogCfg := d.GetChange(h.GetKey())

	if oldLogCfg == nil {
		oldLogCfg = new(schema.Set)
	}
	if newLogCfg == nil {
		newLogCfg = new(schema.Set)
	}

	oldLogSet := oldLogCfg.(*schema.Set)
	newLogSet := newLogCfg.(*schema.Set)

	removeScalyrLogging := oldLogSet.Difference(newLogSet).List()
	addScalyrLogging := newLogSet.Difference(oldLogSet).List()

	// DELETE old Scalyr logging endpoints.
	for _, oRaw := range removeScalyrLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Scalyr logging endpoint removal opts: %#v", opts)

		if err := deleteScalyr(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Scalyr logging endponts.
	for _, nRaw := range addScalyrLogging {
		cfg := nRaw.(map[string]interface{})
		opts := h.buildCreate(cfg, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Scalyr logging addition opts: %#v", opts)

		if err := createScalyr(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *ScalyrServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh Scalyr.
	log.Printf("[DEBUG] Refreshing Scalyr logging endpoints for (%s)", d.Id())
	scalyrList, err := conn.ListScalyrs(&gofastly.ListScalyrsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Scalyr logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	scalyrLogList := flattenScalyr(scalyrList)

	if err := d.Set(h.GetKey(), scalyrLogList); err != nil {
		log.Printf("[WARN] Error setting Scalyr logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createScalyr(conn *gofastly.Client, i *gofastly.CreateScalyrInput) error {
	_, err := conn.CreateScalyr(i)
	return err
}

func deleteScalyr(conn *gofastly.Client, i *gofastly.DeleteScalyrInput) error {
	err := conn.DeleteScalyr(i)

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

func flattenScalyr(scalyrList []*gofastly.Scalyr) []map[string]interface{} {
	var flattened []map[string]interface{}
	for _, s := range scalyrList {
		// Convert logging to a map for saving to state.
		flatScalyr := map[string]interface{}{
			"name":               s.Name,
			"region":             s.Region,
			"token":              s.Token,
			"response_condition": s.ResponseCondition,
			"format":             s.Format,
			"placement":          s.Placement,
			"format_version":     s.FormatVersion,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range flatScalyr {
			if v == "" {
				delete(flatScalyr, k)
			}
		}

		flattened = append(flattened, flatScalyr)
	}

	return flattened
}

func (h *ScalyrServiceAttributeHandler) buildCreate(scalyrMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateScalyrInput {
	df := scalyrMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateScalyrInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              fastly.NullString(df["name"].(string)),
		Region:            fastly.NullString(df["region"].(string)),
		Token:             fastly.NullString(df["token"].(string)),
		Format:            gofastly.NullString(vla.format),
		FormatVersion:     vla.formatVersion,
		Placement:         gofastly.NullString(vla.placement),
		ResponseCondition: gofastly.NullString(vla.responseCondition),
	}
}

func (h *ScalyrServiceAttributeHandler) buildDelete(scalyrMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteScalyrInput {
	df := scalyrMap.(map[string]interface{})

	return &gofastly.DeleteScalyrInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
