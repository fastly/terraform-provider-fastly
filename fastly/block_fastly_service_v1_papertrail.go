package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var papertrailSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this logging setup",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The address of the papertrail service",
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The port of the papertrail service",
			},
			// Optional fields
			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "%h %l %u %t %r %>s",
				Description: "Apache-style string or VCL variables to use for log formatting",
			},
			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition to apply this logging",
			},
			"placement": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Where in the generated VCL the logging call should be placed.",
				ValidateFunc: validateLoggingPlacement(),
			},
		},
	},
}


func processPapertrail(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	os, ns := d.GetChange("papertrail")
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removePapertrail := oss.Difference(nss).List()
	addPapertrail := nss.Difference(oss).List()

	// DELETE old papertrail configurations
	for _, pRaw := range removePapertrail {
		pf := pRaw.(map[string]interface{})
		opts := gofastly.DeletePapertrailInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    pf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Papertrail removal opts: %#v", opts)
		err := conn.DeletePapertrail(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Papertrail
	for _, pRaw := range addPapertrail {
		pf := pRaw.(map[string]interface{})

		opts := gofastly.CreatePapertrailInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              pf["name"].(string),
			Address:           pf["address"].(string),
			Port:              uint(pf["port"].(int)),
			Format:            pf["format"].(string),
			ResponseCondition: pf["response_condition"].(string),
			Placement:         pf["placement"].(string),
		}

		log.Printf("[DEBUG] Create Papertrail Opts: %#v", opts)
		_, err := conn.CreatePapertrail(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}


func readPapertrail(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing Papertrail for (%s)", d.Id())
	papertrailList, err := conn.ListPapertrails(&gofastly.ListPapertrailsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Papertrail for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	pl := flattenPapertrails(papertrailList)

	if err := d.Set("papertrail", pl); err != nil {
		log.Printf("[WARN] Error setting Papertrail for (%s): %s", d.Id(), err)
	}

	return nil
}


func flattenPapertrails(papertrailList []*gofastly.Papertrail) []map[string]interface{} {
	var pl []map[string]interface{}
	for _, p := range papertrailList {
		// Convert Papertrails to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"address":            p.Address,
			"port":               p.Port,
			"format":             p.Format,
			"response_condition": p.ResponseCondition,
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		pl = append(pl, ns)
	}

	return pl
}