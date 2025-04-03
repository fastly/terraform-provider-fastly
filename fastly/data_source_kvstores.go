package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyKVStores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyKVStoresRead,
		Schema: map[string]*schema.Schema{
			"stores": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all KV Stores.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the KV Store.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name for the KV Store.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyKVStoresRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading KV Stores")

	var (
		cursor string
		stores []gofastly.KVStore
	)

	for {
		remoteState, err := conn.ListKVStores(&gofastly.ListKVStoresInput{
			Cursor: cursor,
		})
		if err != nil {
			return diag.Errorf("error fetching KV Stores: %s", err)
		}

		if remoteState != nil {
			stores = append(stores, remoteState.Data...)
			c, ok := remoteState.Meta["next_cursor"]
			if !ok || c == "" || c == cursor {
				break
			}
			cursor = c
		}
	}

	hashBase, _ := json.Marshal(stores)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("stores", flattenDataSourceKVStores(stores)); err != nil {
		return diag.Errorf("error setting stores: %s", err)
	}

	return nil
}

// flattenDataSourceKVStores models data into format suitable for saving to
// Terraform state.
func flattenDataSourceKVStores(remoteState []gofastly.KVStore) []map[string]any {
	result := make([]map[string]any, len(remoteState))
	if len(remoteState) == 0 {
		return result
	}

	for i, resource := range remoteState {
		result[i] = map[string]any{
			"id":   resource.StoreID,
			"name": resource.Name,
		}
	}

	return result
}
