package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServiceDictionaryItemsV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceDictionaryItemsV1Create,
		Read:   resourceServiceDictionaryItemsV1Read,
		Update: resourceServiceDictionaryItemsV1Update,
		Delete: resourceServiceDictionaryItemsV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Id",
			},

			"dictionary_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Dictionary Id",
			},

			"items": {
				Type:        schema.TypeMap,
				Description: "Dictionary Items",
				Optional:    true,
			},
		},
	}
}

func resourceServiceDictionaryItemsV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]interface{})

	for k, v := range items {
		_, err := conn.CreateDictionaryItem(&gofastly.CreateDictionaryItemInput{
			Service:    serviceID,
			Dictionary: dictionaryID,
			ItemKey:    k,
			ItemValue:  v.(string),
		})

		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 409 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, dictionaryID))
	return resourceServiceDictionaryItemsV1Read(d, meta)
}

func resourceServiceDictionaryItemsV1Update(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)

	d.Partial(true)

	if d.HasChange("items") {

		o, n := d.GetChange("items")

		os := o.(map[string]interface{})
		ns := n.(map[string]interface{})

		// Handle Removal
		for k := range os {
			if _, ok := ns[k]; !ok {
				err := conn.DeleteDictionaryItem(&gofastly.DeleteDictionaryItemInput{
					serviceID,
					dictionaryID,
					k,
				})

				if err != nil {
					return fmt.Errorf("Error updating, (removal) dictionary items: service %s, dictionary %s, %#v", serviceID, dictionaryID, err)
				}
			}
		}

		for k, v := range ns {
			// Handle replaces
			if _, ok := os[k]; ok {

				_, err := conn.UpdateDictionaryItem(&gofastly.UpdateDictionaryItemInput{
					serviceID,
					dictionaryID,
					k,
					v.(string),
				})

				if err != nil {
					return fmt.Errorf("Error updating, (update) dictionary items: service %s, dictionary %s, %#v", serviceID, dictionaryID, err)
				}
			}

			// Handle additions
			if _, ok := os[k]; !ok {

				_, err := conn.CreateDictionaryItem(&gofastly.CreateDictionaryItemInput{
					serviceID,
					dictionaryID,
					k,
					v.(string),
				})

				if err != nil {
					return fmt.Errorf("Error updating, (create) dictionary items: service %s, dictionary %s, %#v", serviceID, dictionaryID, err)
				}
			}
		}

		d.SetPartial("items")
	}

	d.Partial(false)

	return resourceServiceDictionaryItemsV1Read(d, meta)
}

func resourceServiceDictionaryItemsV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)

	// TODO Size check
	dictList, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
		Service:    serviceID,
		Dictionary: dictionaryID,
	})
	if err != nil {
		return err
	}

	filteredDictList := filterDictionaryItems(dictList, func(currentDictItem gofastly.DictionaryItem) bool {

		data := d.Get("items")
		items := data.(map[string]interface{})

		if _, ok := items[currentDictItem.ItemKey]; ok {
			return true
		}

		return false
	})

	d.Set("items", flattenDictionaryItems(filteredDictList))
	return nil
}

func resourceServiceDictionaryItemsV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]interface{})

	for k, _ := range items {
		err := conn.DeleteDictionaryItem(&gofastly.DeleteDictionaryItemInput{
			Service:    serviceID,
			Dictionary: dictionaryID,
			ItemKey:    k,
		})

		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 409 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	d.SetId("")
	return nil
}

func filterDictionaryItems(dictList []*gofastly.DictionaryItem, f func(gofastly.DictionaryItem) bool) []*gofastly.DictionaryItem {
	filteredDictList := make([]*gofastly.DictionaryItem, 0)
	for _, item := range dictList {
		if f(*item) {
			filteredDictList = append(filteredDictList, item)
		}
	}

	return filteredDictList
}

func flattenDictionaryItems(dictItemList []*gofastly.DictionaryItem) map[string]string {
	resultList := make(map[string]string)
	for _, currentDictItem := range dictItemList {
		resultList[currentDictItem.ItemKey] = currentDictItem.ItemValue
	}

	return resultList
}
