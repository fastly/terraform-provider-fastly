package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DictionaryServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceDictionary(sa ServiceAttributes) ServiceAttributeDefinition {
	return &DictionaryServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "dictionary",
			serviceAttributes: sa,
		},
	}
}

func (h *DictionaryServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	oldDictVal, newDictVal := d.GetChange(h.GetKey())

	if oldDictVal == nil {
		oldDictVal = new(schema.Set)
	}
	if newDictVal == nil {
		newDictVal = new(schema.Set)
	}

	oldDictSet := oldDictVal.(*schema.Set)
	newDictSet := newDictVal.(*schema.Set)

	remove := oldDictSet.Difference(newDictSet).List()
	add := newDictSet.Difference(oldDictSet).List()

	// Delete removed dictionary configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteDictionaryInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Dictionary Removal opts: %#v", opts)
		err := conn.DeleteDictionary(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new dictionary configurations
	for _, dRaw := range add {
		opts, err := buildDictionary(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Dicitionary: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly Dictionary Addition opts: %#v", opts)
		_, err = conn.CreateDictionary(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *DictionaryServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Dictionaries for (%s)", d.Id())
	dictList, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Dictionaries for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	dict := flattenDictionaries(dictList)

	if err := d.Set(h.GetKey(), dict); err != nil {
		log.Printf("[WARN] Error setting Dictionary for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *DictionaryServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name to refer to this Dictionary",
				},
				// Optional fields
				"dictionary_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Generated dictionary ID",
				},
				"write_only": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Determines if items in the dictionary are readable or not",
				},
			},
		},
	}
	return nil
}

func flattenDictionaries(dictList []*gofastly.Dictionary) []map[string]interface{} {
	var dl []map[string]interface{}
	for _, currentDict := range dictList {

		dictMapString := map[string]interface{}{
			"dictionary_id": currentDict.ID,
			"name":          currentDict.Name,
			"write_only":    currentDict.WriteOnly,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range dictMapString {
			if v == "" {
				delete(dictMapString, k)
			}
		}

		dl = append(dl, dictMapString)
	}

	return dl
}

func buildDictionary(dictMap interface{}) (*gofastly.CreateDictionaryInput, error) {
	df := dictMap.(map[string]interface{})
	opts := gofastly.CreateDictionaryInput{
		Name:      df["name"].(string),
		WriteOnly: gofastly.CBool(df["write_only"].(bool)),
	}

	return &opts, nil
}
