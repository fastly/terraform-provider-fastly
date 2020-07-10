package fastly

import (
	"errors"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type VCLServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceVCL(sa ServiceMetadata) ServiceAttributeDefinition {
	return &VCLServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "vcl",
			serviceMetadata: sa,
		},
	}
}

func (h *VCLServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	// Note: as above with Gzip and S3 logging, we don't utilize the PUT
	// endpoint to update a VCL, we simply destroy it and create a new one.
	oldVCLVal, newVCLVal := d.GetChange(h.GetKey())
	if oldVCLVal == nil {
		oldVCLVal = new(schema.Set)
	}
	if newVCLVal == nil {
		newVCLVal = new(schema.Set)
	}

	oldVCLSet := oldVCLVal.(*schema.Set)
	newVCLSet := newVCLVal.(*schema.Set)

	remove := oldVCLSet.Difference(newVCLSet).List()
	add := newVCLSet.Difference(oldVCLSet).List()

	// Delete removed VCL configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteVCLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Removal opts: %#v", opts)
		err := conn.DeleteVCL(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	// POST new VCL configurations
	for _, dRaw := range add {
		df := dRaw.(map[string]interface{})
		opts := gofastly.CreateVCLInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
			Content: df["content"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Addition opts: %#v", opts)
		_, err := conn.CreateVCL(&opts)
		if err != nil {
			return err
		}

		// if this new VCL is the main
		if df["main"].(bool) {
			opts := gofastly.ActivateVCLInput{
				Service: d.Id(),
				Version: latestVersion,
				Name:    df["name"].(string),
			}
			log.Printf("[DEBUG] Fastly VCL activation opts: %#v", opts)
			_, err := conn.ActivateVCL(&opts)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (h *VCLServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing VCLs for (%s)", d.Id())
	vclList, err := conn.ListVCLs(&gofastly.ListVCLsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up VCLs for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	vl := flattenVCLs(vclList)

	if err := d.Set(h.GetKey(), vl); err != nil {
		log.Printf("[WARN] Error setting VCLs for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *VCLServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A name to refer to this VCL configuration",
				},
				"content": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The contents of this VCL configuration",
				},
				"main": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Should this VCL configuration be the main configuration",
				},
			},
		},
	}
	return nil
}

func flattenVCLs(vclList []*gofastly.VCL) []map[string]interface{} {
	var vl []map[string]interface{}
	for _, vcl := range vclList {
		// Convert VCLs to a map for saving to state.
		vclMap := map[string]interface{}{
			"name":    vcl.Name,
			"content": vcl.Content,
			"main":    vcl.Main,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range vclMap {
			if v == "" {
				delete(vclMap, k)
			}
		}

		vl = append(vl, vclMap)
	}

	return vl
}

func validateVCLs(d *schema.ResourceData) error {
	// TODO: this would be nice to move into a resource/collection validation function, once that is available
	// (see https://github.com/hashicorp/terraform/pull/4348 and https://github.com/hashicorp/terraform/pull/6508)
	vcls, exists := d.GetOk("vcl")
	if !exists {
		return nil
	}

	numberOfMainVCLs, numberOfIncludeVCLs := 0, 0
	for _, vclElem := range vcls.(*schema.Set).List() {
		vcl := vclElem.(map[string]interface{})
		if mainVal, hasMain := vcl["main"]; hasMain && mainVal.(bool) {
			numberOfMainVCLs++
		} else {
			numberOfIncludeVCLs++
		}
	}
	if numberOfMainVCLs == 0 && numberOfIncludeVCLs > 0 {
		return errors.New("if you include VCL configurations, one of them should have main = true")
	}
	if numberOfMainVCLs > 1 {
		return errors.New("you cannot have more than one VCL configuration with main = true")
	}
	return nil
}
