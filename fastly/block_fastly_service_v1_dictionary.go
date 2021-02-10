package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DictionaryServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceDictionary(sa ServiceMetadata) ServiceAttributeDefinition {
	return &DictionaryServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "dictionary",
			serviceMetadata: sa,
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

	oldSet := oldDictVal.(*schema.Set)
	newSet := newDictVal.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		// Use the resource name as the key
		return resource.(map[string]interface{})["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// Delete removed dictionary configurations
	for _, dRaw := range diffResult.Deleted {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteDictionaryInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           df["name"].(string),
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
	for _, dRaw := range diffResult.Added {
		opts, err := buildDictionary(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Dicitionary: %s", err)
			return err
		}
		opts.ServiceID = d.Id()
		opts.ServiceVersion = latestVersion

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
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
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
					Description: "A unique name to identify this dictionary",
				},
				// Optional fields
				"dictionary_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the dictionary",
				},
				"write_only": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "If `true`, the dictionary is a private dictionary, and items are not readable in the UI or via API. Default is `false`. It is important to note that changing this attribute will delete and recreate the dictionary, discard the current items in the dictionary. Using a write-only/private dictionary should only be done if the items are managed outside of Terraform",
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
		WriteOnly: gofastly.Compatibool(df["write_only"].(bool)),
	}

	return &opts, nil
}
