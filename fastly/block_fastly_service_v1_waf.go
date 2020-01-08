package fastly

import (
	"fmt"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// wafSchema the WAF block schema
var wafSchema = &schema.Schema{
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
				Required:    true,
				Description: "The Web Application Firewall's (WAF) prefetch condition",
			},
			"waf_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Web Application Firewall (WAF) ID",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "A flag used to completely disable a Web Application Firewall. This is intended to be used as an emergency.",
			},
		},
	},
}

func processWAF(d *schema.ResourceData, conn *gofastly.Client, v int) error {

	serviceID := d.Id()
	serviceVersion := strconv.Itoa(v)
	oldWAFVal, newWAFVal := d.GetChange("waf")

	if len(newWAFVal.([]interface{})) > 0 {

		wf := newWAFVal.([]interface{})[0].(map[string]interface{})

		var err error
		var waf *gofastly.WAF
		if wafExists(conn, serviceID, serviceVersion, wf["waf_id"].(string)) {
			opts := buildUpdateWAF(wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF update opts: %#v", opts)
			waf, err = conn.UpdateWAF(opts)
		} else {
			opts := buildCreateWAF(wf, serviceID, serviceVersion)
			log.Printf("[DEBUG] Fastly WAF Addition opts: %#v", opts)
			waf, err = conn.CreateWAF(opts)
		}
		if err != nil {
			return err
		}

		if err = processDisabledField(conn, wf, waf); err != nil {
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

func readWAF(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
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
		"disabled":           w.Disabled,
	}

	// prune any empty values that come from the default string value in structs
	for k, v := range WAFMapString {
		if v == "" {
			delete(WAFMapString, k)
		}
	}
	return append(wl, WAFMapString)
}

func processDisabledField(conn *gofastly.Client, wf map[string]interface{}, waf *gofastly.WAF) error {

	if wf["disabled"].(bool) == waf.Disabled {
		return nil
	}

	var err error
	if wf["disabled"].(bool) {
		opts := buildDisableWAF(wf)
		log.Printf("[DEBUG] Fastly WAF disable opts: %#v", opts)
		_, err = conn.DisableWAF(opts)

	} else {
		opts := buildEnableWAF(wf)
		log.Printf("[DEBUG] Fastly WAF enable opts: %#v", opts)
		_, err = conn.EnableWAF(opts)
	}
	if err != nil {
		return err
	}
	return nil
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

func buildDisableWAF(wafMap interface{}) *gofastly.DisableWAFInput {
	df := wafMap.(map[string]interface{})
	opts := gofastly.DisableWAFInput{
		ID: df["waf_id"].(string),
	}
	return &opts
}

func buildEnableWAF(wafMap interface{}) *gofastly.EnableWAFInput {
	df := wafMap.(map[string]interface{})
	opts := gofastly.EnableWAFInput{
		ID: df["waf_id"].(string),
	}
	return &opts
}
