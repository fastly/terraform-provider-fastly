package fastly

import (
	"fmt"
	"strings"

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

	// TODO review this
	d.SetId(fmt.Sprintf("%s/%s", serviceID, dictionaryID))
	return resourceServiceDictionaryItemsV1Read(d, meta)
}

func resourceServiceDictionaryItemsV1Update(d *schema.ResourceData, meta interface{}) error {

	return resourceServiceDictionaryItemsV1Read(d, meta)
}

func resourceServiceDictionaryItemsV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	comp := strings.Split(d.Id(), "/")
	// TODO Size check
	dictList, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
		Service:    comp[0],
		Dictionary: comp[1],
	})
	if err != nil {
		return err
	}

	d.Set("items", flattenDictionaryItems(dictList))

	return nil
}

func resourceServiceDictionaryItemsV1Delete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func flattenDictionaryItems(dictItemList []*gofastly.DictionaryItem) map[string]string {
	resultList := make(map[string]string)
	for _, currentDictItem := range dictItemList {
		resultList[currentDictItem.ItemKey] = currentDictItem.ItemValue
	}

	return resultList
}
