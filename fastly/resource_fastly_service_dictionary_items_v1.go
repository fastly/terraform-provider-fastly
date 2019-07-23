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
				Description: "The service the dictionary belongs to",
			},

			"dictionary_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The dictionary the items belong to",
			},

			"items": {
				Type:         schema.TypeMap,
				Optional:     true,
				Description:  "Key/value pairs that make up an item in the dictionary",
				ValidateFunc: validateDictionaryItems(),
			},
		},
	}
}

func resourceServiceDictionaryItemsV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]interface{})

	var batchDictionaryItems = []*gofastly.BatchDictionaryItem{}

	for key, val := range items {

		batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
			Operation: gofastly.CreateBatchOperation,
			ItemKey:   key,
			ItemValue: val.(string),
		})
	}

	// Process the batch operations
	err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
	if err != nil {
		return fmt.Errorf("Error creating dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
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

		var batchDictionaryItems = []*gofastly.BatchDictionaryItem{}

		o, n := d.GetChange("items")

		os := o.(map[string]interface{})
		ns := n.(map[string]interface{})

		// Handle Removal
		for key := range os {
			if _, ok := ns[key]; !ok {

				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.DeleteBatchOperation,
					ItemKey:   key,
				})
			}
		}

		for key, val := range ns {
			// Handle replaces
			if _, ok := os[key]; ok {

				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.UpdateBatchOperation,
					ItemKey:   key,
					ItemValue: val.(string),
				})
			}

			// Handle additions
			if _, ok := os[key]; !ok {

				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.CreateBatchOperation,
					ItemKey:   key,
					ItemValue: val.(string),
				})

			}
		}

		// Process the batch operations
		err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
		if err != nil {
			return fmt.Errorf("Error updating dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
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

	dictList, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
		Service:    serviceID,
		Dictionary: dictionaryID,
	})
	if err != nil {
		return err
	}

	d.Set("items", flattenDictionaryItems(dictList))
	return nil
}

func resourceServiceDictionaryItemsV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]interface{})

	var batchDictionaryItems = []*gofastly.BatchDictionaryItem{}

	for key, _ := range items {

		batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
			Operation: gofastly.DeleteBatchOperation,
			ItemKey:   key,
		})
	}

	// Process the batch operations
	err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
	if err != nil {
		return fmt.Errorf("Error creating dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
	}

	d.SetId("")
	return nil
}

func flattenDictionaryItems(dictItemList []*gofastly.DictionaryItem) map[string]string {
	resultList := make(map[string]string)
	for _, currentDictItem := range dictItemList {
		resultList[currentDictItem.ItemKey] = currentDictItem.ItemValue
	}

	return resultList
}

func executeBatchDictionaryOperations(conn *gofastly.Client, serviceID, dictionaryID string, batchDictionaryItems []*gofastly.BatchDictionaryItem) error {

	batchSize := gofastly.BatchModifyMaximumOperations

	for i := 0; i < len(batchDictionaryItems); i += batchSize {
		j := i + batchSize
		if j > len(batchDictionaryItems) {
			j = len(batchDictionaryItems)
		}

		err := conn.BatchModifyDictionaryItems(&gofastly.BatchModifyDictionaryItemsInput{
			Service:    serviceID,
			Dictionary: dictionaryID,
			Items:      batchDictionaryItems[i:j],
		})

		if err != nil {
			return err
		}

	}

	return nil
}
