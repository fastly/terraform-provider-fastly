package fastly

import (
	"fmt"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type WAFServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceWAF(sa ServiceMetadata) ServiceAttributeDefinition {
	return &WAFServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "waf",
			serviceMetadata: sa,
		},
	}
}

func (h *WAFServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"response_object": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The Web Application Firewall's (WAF) response object",
				},
				"prefetch_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The Web Application Firewall's (WAF) prefetch condition",
				},
				"waf_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Web Application Firewall (WAF) ID",
				},
			},
		},
	}

	return nil
}

func (h *WAFServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	serviceVersion := strconv.Itoa(latestVersion)
	oldWAFVal, newWAFVal := d.GetChange(h.GetKey())

	if len(newWAFVal.([]interface{})) > 0 {
		wf := newWAFVal.([]interface{})[0].(map[string]interface{})

		var err error
		if wafExists(conn, serviceID, serviceVersion, wf["waf_id"].(string)) {
			opts := buildUpdateWAF(wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF update opts: %#v", opts)
			_, err = conn.UpdateWAF(opts)
		} else {
			opts := buildCreateWAF(wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF Addition opts: %#v", opts)

			_, err = conn.CreateWAF(opts)
		}
		if err != nil {
			return err
		}
	} else if len(oldWAFVal.([]interface{})) > 0 {

		wf := oldWAFVal.([]interface{})[0].(map[string]interface{})

		opts := buildDeleteWAF(wf, serviceVersion)
		log.Printf("[DEBUG] Fastly WAF Removal opts: %#v", opts)
		err := conn.DeleteWAF(opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (h *WAFServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// refresh WAFs
	log.Printf("[DEBUG] Refreshing WAFs for (%s)", d.Id())
	wafList, err := conn.ListWAFs(&gofastly.ListWAFsInput{
		FilterService: d.Id(),
		FilterVersion: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up WAFs for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	waf := flattenWAFs(wafList.Items)

	if err := d.Set("waf", waf); err != nil {
		log.Printf("[WARN] Error setting waf for (%s): %s", d.Id(), err)
	}
	return nil
}

func wafExists(conn *gofastly.Client, s, v, id string) bool {

	_, err := conn.GetWAF(&gofastly.GetWAFInput{
		Service: s,
		Version: v,
		ID:      id,
	})
	if err != nil {
		return false
	}
	return true
}

func flattenWAFs(wafList []*gofastly.WAF) []map[string]interface{} {

	var wl []map[string]interface{}
	if len(wafList) == 0 {
		return wl
	}

	w := wafList[0]
	WAFMapString := map[string]interface{}{
		"waf_id":             w.ID,
		"response_object":    w.Response,
		"prefetch_condition": w.PrefetchCondition,
	}

	// prune any empty values that come from the default string value in structs
	for k, v := range WAFMapString {
		if v == "" {
			delete(WAFMapString, k)
		}
	}
	return append(wl, WAFMapString)
}

func buildCreateWAF(WAFMap interface{}, serviceID string, ServiceVersion string) *gofastly.CreateWAFInput {
	df := WAFMap.(map[string]interface{})
	opts := gofastly.CreateWAFInput{
		Service:           serviceID,
		Version:           ServiceVersion,
		ID:                df["waf_id"].(string),
		PrefetchCondition: df["prefetch_condition"].(string),
		Response:          df["response_object"].(string),
	}
	return &opts
}

func buildDeleteWAF(WAFMap interface{}, ServiceVersion string) *gofastly.DeleteWAFInput {
	df := WAFMap.(map[string]interface{})
	opts := gofastly.DeleteWAFInput{
		ID:      df["waf_id"].(string),
		Version: ServiceVersion,
	}
	return &opts
}

func buildUpdateWAF(wafMap interface{}, serviceID string, ServiceVersion string) *gofastly.UpdateWAFInput {
	df := wafMap.(map[string]interface{})
	opts := gofastly.UpdateWAFInput{
		Service:           serviceID,
		Version:           ServiceVersion,
		ID:                df["waf_id"].(string),
		PrefetchCondition: df["prefetch_condition"].(string),
		Response:          df["response_object"].(string),
	}
	return &opts
}
