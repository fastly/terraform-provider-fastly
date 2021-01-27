package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SumologicServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceSumologic(sa ServiceMetadata) ServiceAttributeDefinition {
	return &SumologicServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "sumologic",
			serviceMetadata: sa,
		},
	}
}

func (h *SumologicServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeSumologic := oss.Difference(nss).List()
	addSumologic := nss.Difference(oss).List()

	// DELETE old sumologic configurations
	for _, pRaw := range removeSumologic {
		sf := pRaw.(map[string]interface{})
		opts := gofastly.DeleteSumologicInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           sf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Sumologic removal opts: %#v", opts)
		err := conn.DeleteSumologic(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Sumologic
	for _, pRaw := range addSumologic {
		sf := pRaw.(map[string]interface{})

		var vla = h.getVCLLoggingAttributes(sf)
		opts := gofastly.CreateSumologicInput{
			ServiceID:         d.Id(),
			ServiceVersion:    latestVersion,
			Name:              sf["name"].(string),
			URL:               sf["url"].(string),
			MessageType:       sf["message_type"].(string),
			Format:            vla.format,
			FormatVersion:     int(uintOrDefault(vla.formatVersion)),
			ResponseCondition: vla.responseCondition,
			Placement:         vla.placement,
		}

		log.Printf("[DEBUG] Create Sumologic Opts: %#v", opts)
		_, err := conn.CreateSumologic(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *SumologicServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Sumologic for (%s)", d.Id())
	sumologicList, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Sumologic for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	sul := flattenSumologics(sumologicList)
	if err := d.Set(h.GetKey(), sul); err != nil {
		log.Printf("[WARN] Error setting Sumologic for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *SumologicServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify this Sumologic endpoint",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The URL to Sumologic collector endpoint",
		},
		// Optional fields
		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`. Default `classic`. See [Fastly's Documentation on Sumologic](https://developer.fastly.com/reference/api/logging/sumologic/)",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t %r %>s",
			Description: "Apache-style string or VCL variables to use for log formatting",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of blockAttributes condition to apply this logging.",
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

func flattenSumologics(sumologicList []*gofastly.Sumologic) []map[string]interface{} {
	var l []map[string]interface{}
	for _, p := range sumologicList {
		// Convert Sumologic to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"url":                p.URL,
			"format":             p.Format,
			"response_condition": p.ResponseCondition,
			"message_type":       p.MessageType,
			"format_version":     int(p.FormatVersion),
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		l = append(l, ns)
	}

	return l
}
