package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceDictionaryItems() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceDictionaryItemsCreate,
		ReadContext:   resourceServiceDictionaryItemsRead,
		UpdateContext: resourceServiceDictionaryItemsUpdate,
		DeleteContext: resourceServiceDictionaryItemsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceDictionaryItemsImport,
		},
		Schema: map[string]*schema.Schema{
			"dictionary_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the dictionary that the items belong to",
			},
			"items": {
				Type:             schema.TypeMap,
				Optional:         true,
				Description:      "A map representing an entry in the dictionary, (key/value)",
				ValidateDiagFunc: validateDictionaryItems(),
				Elem:             schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.HasChange("dictionary_id") && !d.Get("manage_items").(bool)
				},
			},
			"manage_items": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Whether to reapply changes if the state of the items drifts, i.e. if items are managed externally",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the service that the dictionary belongs to",
			},
		},
	}
}

func resourceServiceDictionaryItemsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]any)

	var batchDictionaryItems []*gofastly.BatchDictionaryItem

	for key, val := range items {
		batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
			Operation: gofastly.ToPointer(gofastly.CreateBatchOperation),
			ItemKey:   gofastly.ToPointer(key),
			ItemValue: gofastly.ToPointer(val.(string)),
		})
	}

	// Process the batch operations
	err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
	if err != nil {
		return diag.Errorf("error creating dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, dictionaryID))
	return resourceServiceDictionaryItemsRead(ctx, d, meta)
}

func resourceServiceDictionaryItemsUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)

	if d.HasChange("items") {
		var batchDictionaryItems []*gofastly.BatchDictionaryItem

		o, n := d.GetChange("items")

		os := o.(map[string]any)
		ns := n.(map[string]any)

		// Handle Removal
		for key := range os {
			if _, ok := ns[key]; !ok {
				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
					ItemKey:   gofastly.ToPointer(key),
				})
			}
		}

		for key, val := range ns {
			// Handle replaces
			if _, ok := os[key]; ok {
				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.ToPointer(gofastly.UpdateBatchOperation),
					ItemKey:   gofastly.ToPointer(key),
					ItemValue: gofastly.ToPointer(val.(string)),
				})
			}

			// Handle additions
			if _, ok := os[key]; !ok {
				batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
					Operation: gofastly.ToPointer(gofastly.CreateBatchOperation),
					ItemKey:   gofastly.ToPointer(key),
					ItemValue: gofastly.ToPointer(val.(string)),
				})
			}
		}

		// Process the batch operations
		err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
		if err != nil {
			return diag.Errorf("error updating dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
		}
	}

	return resourceServiceDictionaryItemsRead(ctx, d, meta)
}

func resourceServiceDictionaryItemsRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Print("[DEBUG] Refreshing Dictionary Items Configuration")

	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)

	remoteState, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
		ServiceID:    serviceID,
		DictionaryID: dictionaryID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("items", flattenDictionaryItems(remoteState))
	return diag.FromErr(err)
}

func resourceServiceDictionaryItemsDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	dictionaryID := d.Get("dictionary_id").(string)
	items := d.Get("items").(map[string]any)

	var batchDictionaryItems []*gofastly.BatchDictionaryItem

	for key := range items {
		batchDictionaryItems = append(batchDictionaryItems, &gofastly.BatchDictionaryItem{
			Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
			ItemKey:   gofastly.ToPointer(key),
		})
	}

	// Process the batch operations
	err := executeBatchDictionaryOperations(conn, serviceID, dictionaryID, batchDictionaryItems)
	if err != nil {
		return diag.Errorf("error deleting dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
	}

	d.SetId("")
	return nil
}

func resourceServiceDictionaryItemsImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("invalid id: %s. The ID should be in the format [service_id]/[dictionary_id]", d.Id())
	}

	serviceID := split[0]
	dictionaryID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("error importing dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
	}

	err = d.Set("dictionary_id", dictionaryID)
	if err != nil {
		return nil, fmt.Errorf("error importing dictionary items: service %s, dictionary %s, %s", serviceID, dictionaryID, err)
	}

	return []*schema.ResourceData{d}, nil
}

// flattenDictionaryItems models data into format suitable for saving to Terraform state.
func flattenDictionaryItems(remoteState []*gofastly.DictionaryItem) map[string]string {
	result := make(map[string]string)
	for _, currentDictItem := range remoteState {
		if currentDictItem.ItemKey != nil && currentDictItem.ItemValue != nil {
			result[*currentDictItem.ItemKey] = *currentDictItem.ItemValue
		}
	}
	return result
}

func executeBatchDictionaryOperations(conn *gofastly.Client, serviceID, dictionaryID string, batchDictionaryItems []*gofastly.BatchDictionaryItem) error {
	batchSize := gofastly.BatchModifyMaximumOperations

	for i := 0; i < len(batchDictionaryItems); i += batchSize {
		j := i + batchSize
		if j > len(batchDictionaryItems) {
			j = len(batchDictionaryItems)
		}

		err := conn.BatchModifyDictionaryItems(&gofastly.BatchModifyDictionaryItemsInput{
			ServiceID:    serviceID,
			DictionaryID: dictionaryID,
			Items:        batchDictionaryItems[i:j],
		})
		if err != nil {
			return err
		}
	}

	return nil
}
