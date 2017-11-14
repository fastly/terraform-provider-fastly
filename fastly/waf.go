package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

type terraformWAF struct {
	waf   *gofastly.WAF
	owasp *gofastly.OWASP
}

var (
	wafIDSchema = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The ID assigned to this WAF.",
	}

	wafSchema = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Importer: &schema.ResourceImporter{},
			Schema: map[string]*schema.Schema{
				"id": wafIDSchema,
				"last_push": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Date and time that VCL was last pushed to cache nodes.",
				},
				"owasp": owaspSchema, // OWASP is embedded because there is no way to list or create these apart from an associated WAF
				"prefetch_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Name of the corresponding condition object.",
				},
				"response": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Name of the corresponding response object.",
				},
			},
		},
	}
)

// expandWAF takes input matching the wafSchema and transforms it into a terraformWAF
func expandWAF(wafMap map[string]interface{}) terraformWAF {
	waf := gofastly.WAF{
		ID:       wafMap["id"].(string),
		LastPush: wafMap["last_push"].(string),
	}
	if prefetch, ok := wafMap["prefetch_condtion"]; ok {
		waf.PrefetchCondition = prefetch.(string)
	}
	if response, ok := wafMap["response"]; ok {
		waf.Response = response.(string)
	}

	tf := terraformWAF{waf: &waf}

	owaspSchema := wafMap["owasp"].(*schema.Set).List()
	if len(owaspSchema) == 1 {
		tf.owasp = expandOWASP(owaspSchema[0].(map[string]interface{}))
	}

	return tf
}

// flattenWAFs converts a list of WAFs to a map for saving to state.
func flattenWAFs(wafList []*gofastly.WAF) []map[string]interface{} {
	var wl []map[string]interface{}
	for _, waf := range wafList {
		wafMap := map[string]interface{}{
			"id":                 waf.ID,
			"last_push":          waf.LastPush,
			"prefetch_condition": waf.PrefetchCondition,
			"response":           waf.Response,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range wafMap {
			if v == "" {
				delete(wafMap, k)
			}
		}

		wl = append(wl, wafMap)
	}

	return wl
}

// refreshWAFs queries the fastly API for the defined WAFs associated with the service id
func refreshWAFs(client *gofastly.Client, id string, version int) ([]map[string]interface{}, error) {
	wafList, err := client.ListWAFs(&gofastly.ListWAFsInput{
		Service: id,
		Version: version,
	})
	if err != nil {
		return nil, err
	}

	wl := flattenWAFs(wafList)

	for _, wafSet := range wl {
		owasp, err := client.GetOWASP(&gofastly.GetOWASPInput{Service: id, ID: wafSet["id"].(string)})
		if err != nil {
			return nil, err
		}
		if owasp == nil {
			continue
		}
		owaspSet := schema.NewSet(schema.HashResource(owaspSchema.Elem.(*schema.Resource)), nil)
		owaspSet.Add(flattenOWASP(owasp))
		wafSet["owasp"] = owaspSet
	}

	return wl, nil
}

// updateWAF checks the changes for waf block of the resource service and updates the WAF resources as appropriate.
func updateWAF(client *gofastly.Client, d *schema.ResourceData, version int) error {
	if !d.HasChange("waf") {
		return nil
	}

	oldWAF, newWAF := d.GetChange("waf")
	if oldWAF == nil {
		oldWAF = new(schema.Set)
	}
	if newWAF == nil {
		newWAF = new(schema.Set)
	}

	oldIDs := schema.NewSet(schema.HashSchema(wafIDSchema), nil)
	for _, wafRaw := range oldWAF.(*schema.Set).List() {
		tfWAF := expandWAF(wafRaw.(map[string]interface{}))
		oldIDs.Add(tfWAF.waf.ID)
	}

	newIDs := schema.NewSet(schema.HashSchema(wafIDSchema), nil)
	newIDMap := make(map[string]terraformWAF)
	for i, wafRaw := range newWAF.(*schema.Set).List() {
		tfWAF := expandWAF(wafRaw.(map[string]interface{}))
		if tfWAF.waf.ID == "" {
			tfWAF.waf.ID = fmt.Sprintf("new-%d", i)
		}
		newIDs.Add(tfWAF.waf.ID)
		newIDMap[tfWAF.waf.ID] = tfWAF
	}

	for _, wafID := range oldIDs.Difference(newIDs).List() {
		// There is no OWASP delete, just remove the WAF which contains it
		deleteWAF := &gofastly.DeleteWAFInput{
			Service: d.Id(),
			Version: version,
			ID:      wafID.(string),
		}
		if err := client.DeleteWAF(deleteWAF); err != nil {
			return err
		}
	}

	for _, wafID := range newIDs.Difference(oldIDs).List() {
		tfWAF := newIDMap[wafID.(string)]
		wafCreate := &gofastly.CreateWAFInput{
			Service:           d.Id(),
			Version:           version,
			PrefetchCondition: tfWAF.waf.PrefetchCondition,
			Response:          tfWAF.waf.Response,
		}
		if _, err := client.CreateWAF(wafCreate); err != nil {
			return err
		}

		if tfWAF.owasp != nil {
			owaspCreate := &gofastly.CreateOWASPInput{
				Service: d.Id(),
				ID:      wafID.(string),
				Type:    "owasp",
			}
			if _, err := client.CreateOWASP(owaspCreate); err != nil {
				return err
			}
		}
	}

	for _, wafID := range oldIDs.Intersection(newIDs).List() {
		tfWAF := newIDMap[wafID.(string)]
		wafUpdate := &gofastly.UpdateWAFInput{
			Service:           d.Id(),
			Version:           version,
			PrefetchCondition: tfWAF.waf.PrefetchCondition,
			Response:          tfWAF.waf.Response,
		}
		if _, err := client.UpdateWAF(wafUpdate); err != nil {
			return err
		}

		if _, err := updateOWASP(client, d.Id(), wafID.(string), tfWAF.owasp); err != nil {
			return err
		}
	}

	return nil
}
