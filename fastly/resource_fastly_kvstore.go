package fastly

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
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
				Description: "Allow the KV Store to be deleted, even if it contains entries. Defaults to false.",
			},
			// TODO: Move values to constants inside of go-fastly.
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The regional location of the KV Store. Valid values are `US`, `EU`, `ASIA`, and `AUS`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"US", "EU", "ASIA", "AUS"},
					false,
				)),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to identify the KV Store. It is important to note that changing this attribute will delete and recreate the KV Store, and discard the current entries. You MUST first delete the associated resource_link block from your service before modifying this field.",
				ForceNew:    true,
			},
		},
	}
}

func resourceFastlyKVStoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.CreateKVStoreInput{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("location"); ok {
		input.Location = v.(string)
	}

	log.Printf("[DEBUG] CREATE: KV Store input: %#v", input)

	store, err := conn.CreateKVStore(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: `id` is exposed as a read-only attribute.
	d.SetId(store.StoreID)

	return nil
}

func resourceFastlyKVStoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetKVStoreInput{
		StoreID: d.Id(),
	}

	log.Printf("[DEBUG] REFRESH: KV Store input: %#v", input)

	store, err := conn.GetKVStore(ctx, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] No KV Store found '%s'", d.Id())
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

// NOTE: There is no UPDATE endpoint for KV Stores.
func resourceFastlyKVStoreUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourceFastlyKVStoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	if !d.Get("force_destroy").(bool) {
		mayDelete, err := isKVStoreEmpty(ctx, d.Id(), conn)
		if err != nil {
			return diag.FromErr(err)
		}

		if !mayDelete {
			return diag.FromErr(fmt.Errorf("cannot delete KV Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", d.Id()))
		}
	}

	// IMPORTANT: We must delete all keys first before we can delete the store.
	p := conn.NewListKVStoreKeysPaginator(ctx, &gofastly.ListKVStoreKeysInput{
		StoreID: d.Id(),
	})
	for p.Next() {
		keys := p.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			err := conn.DeleteKVStoreKey(ctx, &gofastly.DeleteKVStoreKeyInput{
				StoreID: d.Id(),
				Key:     key,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("error during KV Store key cleanup: %w", err))
			}
		}
	}
	if err := p.Err(); err != nil {
		return diag.FromErr(fmt.Errorf("error during KV Store cleanup pagination: %w", err))
	}

	input := gofastly.DeleteKVStoreInput{
		StoreID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: KV Store input: %#v", input)

	err := conn.DeleteKVStore(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func isKVStoreEmpty(ctx context.Context, storeID string, conn *gofastly.Client) (bool, error) {
	keys, err := conn.ListKVStoreKeys(ctx, &gofastly.ListKVStoreKeysInput{
		StoreID: storeID,
	})
	if err != nil {
		return false, err
	}
	return len(keys.Data) == 0, nil
}
