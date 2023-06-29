package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyConfigStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyConfigStoreCreate,
		ReadContext:   resourceFastlyConfigStoreRead,
		UpdateContext: resourceFastlyConfigStoreUpdate,
		DeleteContext: resourceFastlyConfigStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"force_destroy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Allow the Config Store to be deleted, even if it contains entries. Defaults to false.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to identify the Config Store. It is important to note that changing this attribute will delete and recreate the Config Store, and discard the current entries. You MUST first delete the associated resource_link block from your service before modifying this field.",
				ForceNew:    true,
			},
		},
	}
}

func resourceFastlyConfigStoreCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.CreateConfigStoreInput{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] CREATE: Config Store input: %#v", input)

	store, err := conn.CreateConfigStore(input)
	if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: `id` is exposed as a read-only attribute.
	d.SetId(store.ID)

	return nil
}

func resourceFastlyConfigStoreRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetConfigStoreInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] REFRESH: Config Store input: %#v", input)

	store, err := conn.GetConfigStore(input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] No Config Store found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	err = d.Set("name", store.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// NOTE: There is no UPDATE endpoint for Config Stores.
// A change in the name will result in a delete and recreate.
func resourceFastlyConfigStoreUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourceFastlyConfigStoreDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	if !d.Get("force_destroy").(bool) {
		mayDelete, err := isConfigStoreEmpty(d.Id(), conn)
		if err != nil {
			return diag.FromErr(err)
		}

		if !mayDelete {
			return diag.FromErr(fmt.Errorf("cannot delete Config Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", d.Id()))
		}
	}

	items, err := conn.ListConfigStoreItems(&gofastly.ListConfigStoreItemsInput{
		StoreID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// IMPORTANT: We must delete all keys first before we can delete the store.
	for _, item := range items {
		err := conn.DeleteConfigStoreItem(&gofastly.DeleteConfigStoreItemInput{
			StoreID: item.StoreID,
			Key:     item.Key,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error during Config Store key cleanup: %w", err))
		}
	}

	input := gofastly.DeleteConfigStoreInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: Config Store input: %#v", input)

	err = conn.DeleteConfigStore(&input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func isConfigStoreEmpty(storeID string, conn *gofastly.Client) (bool, error) {
	items, err := conn.ListConfigStoreItems(&gofastly.ListConfigStoreItemsInput{
		StoreID: storeID,
	})
	if err != nil {
		return false, err
	}
	return len(items) == 0, nil
}
