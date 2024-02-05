package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyConfigStoreEntries() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyConfigStoreEntriesCreate,
		ReadContext:   resourceFastlyConfigStoreEntriesRead,
		UpdateContext: resourceFastlyConfigStoreEntriesUpdate,
		DeleteContext: resourceFastlyConfigStoreEntriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigStoreEntriesImport,
		},
		Schema: map[string]*schema.Schema{
			"entries": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "A map representing an entry in the Config Store, (key/value)",
				Elem:        schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress the diff unless the user wishes Terraform to manage the entries.
					return !d.HasChange("store_id") && !d.Get("manage_entries").(bool)
				},
			},
			"manage_entries": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Have Terraform manage the entries (default: false). If set to `true` Terraform will remove any entries that were added externally from the config seeded values.",
			},
			"store_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "An alphanumeric string identifying the Config Store.",
			},
		},
	}
}

func resourceFastlyConfigStoreEntriesCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	entries := d.Get("entries").(map[string]any)
	storeID := d.Get("store_id").(string)

	var batchEntries []*gofastly.BatchConfigStoreItem

	for key, val := range entries {
		batchEntries = append(batchEntries, &gofastly.BatchConfigStoreItem{
			Operation: gofastly.CreateBatchOperation,
			ItemKey:   key,
			ItemValue: val.(string),
		})
	}

	log.Printf("[DEBUG] CREATE: Config Store Entries")

	err := executeBatchConfigStoreOperations(conn, storeID, batchEntries)
	if err != nil {
		return diag.Errorf("error creating Config Store (%s) entries: %s", storeID, err)
	}

	// NOTE: `id` is exposed as a read-only attribute.
	d.SetId(fmt.Sprintf("%s/entries", storeID))

	return nil
}

func resourceFastlyConfigStoreEntriesRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] REFRESH: Config Store Entries")

	storeID := d.Get("store_id").(string)

	remoteState, err := conn.ListConfigStoreItems(&gofastly.ListConfigStoreItemsInput{
		StoreID: storeID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("entries", flattenConfigStoreEntries(remoteState))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyConfigStoreEntriesUpdate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	storeID := d.Get("store_id").(string)

	log.Printf("[DEBUG] UPDATE: Config Store Entries")

	if d.HasChange("entries") {
		var batchEntries []*gofastly.BatchConfigStoreItem

		o, n := d.GetChange("entries")

		om := o.(map[string]any)
		nm := n.(map[string]any)

		// Deletions
		for key := range om {
			if _, ok := nm[key]; !ok {
				batchEntries = append(batchEntries, &gofastly.BatchConfigStoreItem{
					Operation: gofastly.DeleteBatchOperation,
					ItemKey:   key,
				})
			}
		}

		for key, val := range nm {
			// Updates
			if _, ok := om[key]; ok {
				batchEntries = append(batchEntries, &gofastly.BatchConfigStoreItem{
					Operation: gofastly.UpdateBatchOperation,
					ItemKey:   key,
					ItemValue: val.(string),
				})
			}

			// Additions
			if _, ok := om[key]; !ok {
				batchEntries = append(batchEntries, &gofastly.BatchConfigStoreItem{
					Operation: gofastly.CreateBatchOperation,
					ItemKey:   key,
					ItemValue: val.(string),
				})
			}
		}

		err := executeBatchConfigStoreOperations(conn, storeID, batchEntries)
		if err != nil {
			return diag.Errorf("error updating Config Store (%s) elements: %s", storeID, err)
		}
	}

	return nil
}

func resourceFastlyConfigStoreEntriesDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	entries := d.Get("entries").(map[string]any)
	storeID := d.Get("store_id").(string)

	log.Printf("[DEBUG] DELETE: Config Store Entries")

	var batchEntries []*gofastly.BatchConfigStoreItem

	for key := range entries {
		batchEntries = append(batchEntries, &gofastly.BatchConfigStoreItem{
			Operation: gofastly.DeleteBatchOperation,
			ItemKey:   key,
		})
	}

	err := executeBatchConfigStoreOperations(conn, storeID, batchEntries)
	if err != nil {
		return diag.Errorf("error deleting Config Store (%s) entries: %s", storeID, err)
	}

	d.SetId("")

	return nil
}

// flattenConfigStoreEntries models data into format suitable for saving to
// Terraform state.
func flattenConfigStoreEntries(remoteState []*gofastly.ConfigStoreItem) map[string]string {
	result := make(map[string]string)
	for _, currentEntry := range remoteState {
		result[currentEntry.Key] = currentEntry.Value
	}
	return result
}

// executeBatchConfigStoreOperations is called from with the Create, Update and
// Delete methods.
func executeBatchConfigStoreOperations(conn *gofastly.Client, storeID string, batchEntries []*gofastly.BatchConfigStoreItem) error {
	batchSize := gofastly.BatchModifyMaximumOperations

	for i := 0; i < len(batchEntries); i += batchSize {
		j := i + batchSize
		if j > len(batchEntries) {
			j = len(batchEntries)
		}

		err := conn.BatchModifyConfigStoreItems(&gofastly.BatchModifyConfigStoreItemsInput{
			StoreID: storeID,
			Items:   batchEntries[i:j],
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceConfigStoreEntriesImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("invalid id: %s. The ID should be in the format [store_id]/entries", d.Id())
	}

	storeID := split[0]

	err := d.Set("store_id", storeID)
	if err != nil {
		return nil, fmt.Errorf("error setting Config Store ID (%s): %s", storeID, err)
	}

	return []*schema.ResourceData{d}, nil
}
