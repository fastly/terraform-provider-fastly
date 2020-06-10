package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

)

var sumologicSchema = &schema.Schema{
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
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to POST to.",
			},
			// Optional fields
			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "%h %l %u %t %r %>s",
				Description: "Apache-style string or VCL variables to use for log formatting",
			},
			"format_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
				ValidateFunc: validateLoggingFormatVersion(),
			},
			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition to apply this logging.",
			},
			"message_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "classic",
				Description:  "How the message should be formatted.",
				ValidateFunc: validateLoggingMessageType(),
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


func processSumologic(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	os, ns := d.GetChange("sumologic")
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
			Service: d.Id(),
			Version: latestVersion,
			Name:    sf["name"].(string),
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
		opts := gofastly.CreateSumologicInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              sf["name"].(string),
			URL:               sf["url"].(string),
			Format:            sf["format"].(string),
			FormatVersion:     sf["format_version"].(int),
			ResponseCondition: sf["response_condition"].(string),
			MessageType:       sf["message_type"].(string),
			Placement:         sf["placement"].(string),
		}

		log.Printf("[DEBUG] Create Sumologic Opts: %#v", opts)
		_, err := conn.CreateSumologic(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}


func readSumologic(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing Sumologic for (%s)", d.Id())
	sumologicList, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Sumologic for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	sul := flattenSumologics(sumologicList)
	if err := d.Set("sumologic", sul); err != nil {
		log.Printf("[WARN] Error setting Sumologic for (%s): %s", d.Id(), err)
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