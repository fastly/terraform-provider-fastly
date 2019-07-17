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

	var batchItems = []gofastly.BatchDictionaryItem{}

	for key, val := range items {

		batchItems = append(batchItems, gofastly.BatchDictionaryItem{
			Operation: gofastly.Create,
			ItemKey:   key,
			ItemValue: val.(string),
		})
	}

	err := conn.BatchModifyDictionaryItems(&gofastly.BatchModifyDictionaryItemsInput{
		Service:    serviceID,
		Dictionary: dictionaryID,
		Items:      batchItems,
	})

	if err != nil {
		return err
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

		var batchDictionaryItems = []gofastly.BatchDictionaryItem{}

		o, n := d.GetChange("items")

		os := o.(map[string]interface{})
		ns := n.(map[string]interface{})

		// Handle Removal
		for key := range os {
			if _, ok := ns[key]; !ok {

				batchDictionaryItems = append(batchDictionaryItems, gofastly.BatchDictionaryItem{
					Operation: gofastly.Delete,
					ItemKey:   key,
				})
			}
		}

		for key, val := range ns {
			// Handle replaces
			if _, ok := os[key]; ok {

				batchDictionaryItems = append(batchDictionaryItems, gofastly.BatchDictionaryItem{
					Operation: gofastly.Update,
					ItemKey:   key,
					ItemValue: val.(string),
				})
			}

			// Handle additions
			if _, ok := os[key]; !ok {

				batchDictionaryItems = append(batchDictionaryItems, gofastly.BatchDictionaryItem{
					Operation: gofastly.Create,
					ItemKey:   key,
					ItemValue: val.(string),
				})

			}
		}

		err := conn.BatchModifyDictionaryItems(&gofastly.BatchModifyDictionaryItemsInput{
			Service:    serviceID,
			Dictionary: dictionaryID,
			Items:      batchDictionaryItems,
		})

		if err != nil {
			return fmt.Errorf("Error updating dictionary items: service %s, dictionary %s, %#v", serviceID, dictionaryID, err)
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

	var batchItems = []gofastly.BatchDictionaryItem{}

	for key, _ := range items {

		batchItems = append(batchItems, gofastly.BatchDictionaryItem{
			Operation: gofastly.Delete,
			ItemKey:   key,
		})
	}

	err := conn.BatchModifyDictionaryItems(&gofastly.BatchModifyDictionaryItemsInput{
		Service:    serviceID,
		Dictionary: dictionaryID,
		Items:      batchItems,
	})

	if err != nil {
		return err
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
