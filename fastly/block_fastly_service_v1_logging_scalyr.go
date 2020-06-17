package fastly

import (
	"fmt"
	"log"

	"github.com/fastly/go-fastly/fastly"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var scalyrloggingSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
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
			},

			// Optional
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "US",
				Description: "The region that log data will be sent to. One of US or EU. Defaults to US if undefined.",
			},

			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Apache style log formatting.",
			},

			"format_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
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
				Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
			},
		},
	},
}

func processScalyr(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	serviceID := d.Id()
	oldLogCfg, newLogCfg := d.GetChange("logging_scalyr")

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
		opts := buildDeleteScalyr(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Scalyr logging endpoint removal opts: %#v", opts)

		if err := deleteScalyr(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Scalyr logging endponts.
	for _, nRaw := range addScalyrLogging {
		cfg := nRaw.(map[string]interface{})
		opts := buildCreateScalyr(cfg, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Scalyr logging addition opts: %#v", opts)

		if err := createScalyr(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func readScalyr(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
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

	if err := d.Set("logging_scalyr", scalyrLogList); err != nil {
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

func buildCreateScalyr(scalyrMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateScalyrInput {
	df := scalyrMap.(map[string]interface{})

	return &gofastly.CreateScalyrInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              fastly.NullString(df["name"].(string)),
		Region:            fastly.NullString(df["region"].(string)),
		Token:             fastly.NullString(df["token"].(string)),
		Format:            fastly.NullString(df["format"].(string)),
		FormatVersion:     fastly.Uint(uint(df["format_version"].(int))),
		Placement:         fastly.NullString(df["placement"].(string)),
		ResponseCondition: fastly.NullString(df["response_condition"].(string)),
	}
}

func buildDeleteScalyr(scalyrMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteScalyrInput {
	df := scalyrMap.(map[string]interface{})

	return &gofastly.DeleteScalyrInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
