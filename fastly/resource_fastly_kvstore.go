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
			"resource_link": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Resource you want to link the KV Store to.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_version": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"store_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the KV Store.",
			},
		},
	}
}

func resourceFastlyKVStoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.CreateKVStoreInput{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] Fastly KV Store Addition input: %#v", input)
	store, err := conn.CreateKVStore(input)
	if err != nil {
		return diag.FromErr(err)
	}

	resources := d.Get("resource_link").(*schema.Set).List()
	for _, v := range resources {
		resource := v.(map[string]any)
		sid := resource["service_id"].(string)
		sv := resource["service_version"].(int)
		log.Printf("[DEBUG] Fastly KV Store link to resource: %s (version: %d)", sid, sv)
		_, err = conn.CreateResource(&gofastly.CreateResourceInput{
			ServiceID:      sid,
			ServiceVersion: sv,
			Name:           gofastly.String(store.Name),
			ResourceID:     gofastly.String(store.ID),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(store.ID)

	return resourceFastlyKVStoreRead(ctx, d, meta)
}

func resourceFastlyKVStoreRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing KV Store for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	store, err := conn.GetKVStore(&gofastly.GetKVStoreInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("store_id", store.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", store.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: check if we need to reset force_destroy
	// TODO: update all resource links for this service (i.e. resources_links)

	return nil
}

func resourceFastlyKVStoreUpdate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// TODO: Keep resource_links in-sync.
	return nil
}

func resourceFastlyKVStoreDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	storeID := d.Get("store_id").(string)

	if !d.Get("force_destroy").(bool) {
		mayDelete, err := isKVStoreEmpty(storeID, conn)
		if err != nil {
			return diag.FromErr(err)
		}

		if !mayDelete {
			return diag.FromErr(fmt.Errorf("cannot delete KV Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", storeID))
		}
	}

	input := gofastly.DeleteKVStoreInput{
		ID: storeID,
	}

	log.Printf("[DEBUG] Fastly KV Store Removal input: %#v", input)
	err := conn.DeleteKVStore(&input)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		// NOTE: If the ID provided by the user is unrecognised, fail silently.
		if errRes.StatusCode != 404 {
			return diag.FromErr(err)
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func isKVStoreEmpty(dictID string, conn *gofastly.Client) (bool, error) {
	keys, err := conn.ListKVStoreKeys(&gofastly.ListKVStoreKeysInput{
		ID: dictID,
	})
	if err != nil {
		return false, err
	}

	return len(keys.Data) == 0, nil
}
