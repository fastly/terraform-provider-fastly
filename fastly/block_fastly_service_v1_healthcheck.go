package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var healthcheckSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name to refer to this healthcheck",
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Which host to check",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to check",
			},
			// optional fields
			"check_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5000,
				Description: "How often to run the healthcheck in milliseconds",
			},
			"expected_response": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     200,
				Description: "The status code expected from the host",
			},
			"http_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1.1",
				Description: "Whether to use version 1.0 or 1.1 HTTP",
			},
			"initial": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2,
				Description: "When loading a config, the initial number of probes to be seen as OK",
			},
			"method": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "HEAD",
				Description: "Which HTTP method to use",
			},
			"threshold": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "How many healthchecks must succeed to be considered healthy",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     500,
				Description: "Timeout in milliseconds",
			},
			"window": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "The number of most recent healthcheck queries to keep for this healthcheck",
			},
		},
	},
}

func processHealthCheck(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	oh, nh := d.GetChange("healthcheck")
	if oh == nil {
		oh = new(schema.Set)
	}
	if nh == nil {
		nh = new(schema.Set)
	}

	ohs := oh.(*schema.Set)
	nhs := nh.(*schema.Set)
	removeHealthCheck := ohs.Difference(nhs).List()
	addHealthCheck := nhs.Difference(ohs).List()

	// DELETE old healthcheck configurations
	for _, hRaw := range removeHealthCheck {
		hf := hRaw.(map[string]interface{})
		opts := gofastly.DeleteHealthCheckInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    hf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Healthcheck removal opts: %#v", opts)
		err := conn.DeleteHealthCheck(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Healthcheck
	for _, hRaw := range addHealthCheck {
		hf := hRaw.(map[string]interface{})

		opts := gofastly.CreateHealthCheckInput{
			Service:          d.Id(),
			Version:          latestVersion,
			Name:             hf["name"].(string),
			Host:             hf["host"].(string),
			Path:             hf["path"].(string),
			CheckInterval:    uint(hf["check_interval"].(int)),
			ExpectedResponse: uint(hf["expected_response"].(int)),
			HTTPVersion:      hf["http_version"].(string),
			Initial:          uint(hf["initial"].(int)),
			Method:           hf["method"].(string),
			Threshold:        uint(hf["threshold"].(int)),
			Timeout:          uint(hf["timeout"].(int)),
			Window:           uint(hf["window"].(int)),
		}

		log.Printf("[DEBUG] Create Healthcheck Opts: %#v", opts)
		_, err := conn.CreateHealthCheck(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func readHealthCheck(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing Healthcheck for (%s)", d.Id())
	healthcheckList, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Healthcheck for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	hcl := flattenHealthchecks(healthcheckList)

	if err := d.Set("healthcheck", hcl); err != nil {
		log.Printf("[WARN] Error setting Healthcheck for (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenHealthchecks(healthcheckList []*gofastly.HealthCheck) []map[string]interface{} {
	var hl []map[string]interface{}
	for _, h := range healthcheckList {
		// Convert HealthChecks to a map for saving to state.
		nh := map[string]interface{}{
			"name":              h.Name,
			"host":              h.Host,
			"path":              h.Path,
			"check_interval":    h.CheckInterval,
			"expected_response": h.ExpectedResponse,
			"http_version":      h.HTTPVersion,
			"initial":           h.Initial,
			"method":            h.Method,
			"threshold":         h.Threshold,
			"timeout":           h.Timeout,
			"window":            h.Window,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}

	return hl
}
