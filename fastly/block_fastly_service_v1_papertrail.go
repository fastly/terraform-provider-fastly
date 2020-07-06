package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type PaperTrailServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServicePaperTrail(sa ServiceAttributes) ServiceAttributeDefinition {
	return &PaperTrailServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "papertrail",
			serviceAttributes: sa,
		},
	}
}

func (h *PaperTrailServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
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

		var vla = h.getVCLLoggingAttributes(pf)
		opts := gofastly.CreatePapertrailInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              pf["name"].(string),
			Address:           pf["address"].(string),
			Port:              uint(pf["port"].(int)),
			Format:            vla.format,
			ResponseCondition: vla.responseCondition,
			Placement:         vla.placement,
		}

		log.Printf("[DEBUG] Create Papertrail Opts: %#v", opts)
		_, err := conn.CreatePapertrail(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *PaperTrailServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Papertrail for (%s)", d.Id())
	papertrailList, err := conn.ListPapertrails(&gofastly.ListPapertrailsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Papertrail for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	pl := flattenPapertrails(papertrailList)

	if err := d.Set(h.GetKey(), pl); err != nil {
		log.Printf("[WARN] Error setting Papertrail for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *PaperTrailServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
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
	}

	if h.GetServiceAttributes().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t %r %>s",
			Description: "Apache-style string or VCL variables to use for log formatting",
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of blockAttributes condition to apply this logging",
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
