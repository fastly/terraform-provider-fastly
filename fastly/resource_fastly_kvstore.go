package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyKVStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyKVStoreCreate,
		ReadContext:   resourceFastlyKVStoreRead,
		UpdateContext: resourceFastlyKVStoreUpdate,
		DeleteContext: resourceFastlyKVStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"force_destroy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Allow the KV store to be deleted, even if it contains entries. Defaults to false.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to identify the KV Store. It is important to note that changing this attribute will delete and recreate the KV Store, and discard the current entries.",
			},
		},
	}
}

func resourceFastlyKVStoreCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.CreateKVStoreInput{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] CREATE: KV Store input: %#v", input)

	store, err := conn.CreateKVStore(input)
	if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: `id` is exposed as a read-only attribute.
	d.SetId(store.ID)

	return nil
}

func resourceFastlyKVStoreRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetKVStoreInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] REFRESH: KV Store input: %#v", input)

	store, err := conn.GetKVStore(input)
	if err != nil {
		log.Printf("[WARN] No KV Store found '%s'", d.Id())
		d.SetId("")
		return diag.FromErr(err)
	}

	err = d.Set("name", store.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// NOTE: There is no UPDATE endpoint for KV Stores.
func resourceFastlyKVStoreUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourceFastlyKVStoreDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	if !d.Get("force_destroy").(bool) {
		mayDelete, err := isKVStoreEmpty(d.Id(), conn)
		if err != nil {
			return diag.FromErr(err)
		}

		if !mayDelete {
			return diag.FromErr(fmt.Errorf("cannot delete KV Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", d.Id()))
		}
	}

	input := gofastly.DeleteKVStoreInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: KV Store input: %#v", input)

	err := conn.DeleteKVStore(&input)
	if err != nil {
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			// If error is because the resource looks to already be deleted (i.e. 404),
			// then skip returning the error and allow it to fail silently.
			if errRes.StatusCode != 404 {
				return diag.FromErr(err)
			}
		}
	}

	return diag.FromErr(err)
}

func isKVStoreEmpty(storeID string, conn *gofastly.Client) (bool, error) {
	keys, err := conn.ListKVStoreKeys(&gofastly.ListKVStoreKeysInput{
		ID: storeID,
	})
	if err != nil {
		return false, err
	}
	return len(keys.Data) == 0, nil
}
