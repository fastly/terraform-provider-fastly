package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// KVStoreServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type KVStoreServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceKVStore returns a new resource.
func NewServiceKVStore(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&KVStoreServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "kv_store",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *KVStoreServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *KVStoreServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
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
				"store_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the KV Store",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *KVStoreServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, config map[string]any, serviceVersion int, conn *gofastly.Client) error {
	input := &gofastly.CreateKVStoreInput{
		Name: config["name"].(string),
	}

	log.Printf("[DEBUG] Fastly KV Store Addition input: %#v", input)
	_, err := conn.CreateKVStore(input)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *KVStoreServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing KV Stores for (%s)", d.Id())

		kvs := []gofastly.KVStore{}

		var (
			cursor string
			ok     bool
		)

		for {
			stores, err := conn.ListKVStores(&gofastly.ListKVStoresInput{
				Cursor: cursor,
			})
			if err != nil {
				log.Fatal(err)
			}

			kvs = append(kvs, stores.Data...)

			cursor, ok = stores.Meta["next_cursor"]
			if !ok {
				break
			}
		}

		stores := flattenKVStores(kvs)

		// Match up force_destroy on each KV Store from schema.ResourceData to avoid d.Set overwriting it with null
		for _, remoteStore := range stores {
			for _, v := range localState {
				localStore := v.(map[string]any)
				if remoteStore["name"] == localStore["name"] {
					remoteStore["force_destroy"] = localStore["force_destroy"]
					break
				}
			}
		}

		if err := d.Set(h.GetKey(), stores); err != nil {
			log.Printf("[WARN] Error setting KV Store for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *KVStoreServiceAttributeHandler) Update(_ context.Context, _ *schema.ResourceData, _, _ map[string]any, _ int, _ *gofastly.Client) error {
	return nil
}

// Delete deletes the resource.
func (h *KVStoreServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	storeID := resource["store_id"].(string)

	if !resource["force_destroy"].(bool) {
		mayDelete, err := isKVStoreEmpty(storeID, conn)
		if err != nil {
			return err
		}

		if !mayDelete {
			return fmt.Errorf("cannot delete KV Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", resource["store_id"].(string))
		}
	}

	input := gofastly.DeleteKVStoreInput{
		ID: storeID,
	}

	log.Printf("[DEBUG] Fastly KV Store Removal opts: %#v", input)
	err := conn.DeleteKVStore(&input)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		// NOTE: If the ID provided by the user is unrecognised, fail silently.
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenKVStores models data into format suitable for saving to Terraform state.
func flattenKVStores(remoteState []gofastly.KVStore) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"store_id": resource.ID,
			"name":     resource.Name,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
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
