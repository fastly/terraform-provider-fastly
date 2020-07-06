package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SplunkServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceSplunk(sa ServiceAttributes) ServiceAttributeDefinition {
	return &SplunkServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "splunk",
			serviceAttributes: sa,
		},
	}
}

func (h *SplunkServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)

	remove := oss.Difference(nss).List()
	add := nss.Difference(oss).List()

	// DELETE old Splunk logging configurations
	for _, sRaw := range remove {
		sf := sRaw.(map[string]interface{})
		opts := gofastly.DeleteSplunkInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    sf["name"].(string),
		}

		log.Printf("[DEBUG] Splunk removal opts: %#v", opts)
		err := conn.DeleteSplunk(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Splunk configurations
	for _, sRaw := range add {
		sf := sRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("splunk")` in this case.
		if v, ok := sf["name"]; !ok || v.(string) == "" {
			continue
		}

		var vla = h.getVCLLoggingAttributes(sf)
		opts := gofastly.CreateSplunkInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              sf["name"].(string),
			URL:               sf["url"].(string),
			Token:             sf["token"].(string),
			TLSHostname:       sf["tls_hostname"].(string),
			TLSCACert:         sf["tls_ca_cert"].(string),
			Format:            vla.format,
			FormatVersion:     vla.formatVersion,
			ResponseCondition: vla.responseCondition,
			Placement:         vla.placement,
		}

		log.Printf("[DEBUG] Splunk create opts: %#v", opts)
		_, err := conn.CreateSplunk(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *SplunkServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Splunks for (%s)", d.Id())
	splunkList, err := conn.ListSplunks(&gofastly.ListSplunksInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Splunks for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	spl := flattenSplunks(splunkList)

	if err := d.Set(h.GetKey(), spl); err != nil {
		log.Printf("[WARN] Error setting Splunks for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *SplunkServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Splunk logging endpoint",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Splunk URL to stream logs to",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_TOKEN", ""),
			Description: "The Splunk token to be used for authentication",
			Sensitive:   true,
		},
		// Optional fields
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name or blockAttributes Subject Alternative Name (SAN).",
		},
		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_CA_CERT", ""),
			Description: "A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SPLUNK_CA_CERT`.",
		},
	}

	if h.GetServiceAttributes().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t \"%r\" %>s %b",
			Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
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

func flattenSplunks(splunkList []*gofastly.Splunk) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, s := range splunkList {
		// Convert Splunk to a map for saving to state.
		nbs := map[string]interface{}{
			"name":               s.Name,
			"url":                s.URL,
			"format":             s.Format,
			"format_version":     s.FormatVersion,
			"response_condition": s.ResponseCondition,
			"placement":          s.Placement,
			"token":              s.Token,
			"tls_hostname":       s.TLSHostname,
			"tls_ca_cert":        s.TLSCACert,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nbs {
			if v == "" {
				delete(nbs, k)
			}
		}

		sl = append(sl, nbs)
	}

	return sl
}
