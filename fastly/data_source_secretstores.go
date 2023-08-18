package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlySecretStores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlySecretStoresRead,
		Schema: map[string]*schema.Schema{
			"stores": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Secrets Stores.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the Secrets Store.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name for the Secrets Store.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlySecretStoresRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading Secrets Stores")

	var (
		cursor string
		stores []gofastly.SecretStore
	)

	for {
		remoteState, err := conn.ListSecretStores(&gofastly.ListSecretStoresInput{
			Cursor: cursor,
		})
		if err != nil {
			return diag.Errorf("error fetching Secrets Stores: %s", err)
		}

		if remoteState != nil {
			stores = append(stores, remoteState.Data...)
			c := remoteState.Meta.NextCursor
			if c == "" || c == cursor {
				break
			}
			cursor = c
		}
	}

	hashBase, _ := json.Marshal(stores)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("stores", flattenDataSourceSecretStores(stores)); err != nil {
		return diag.Errorf("error setting stores: %s", err)
	}

	return nil
}

// flattenDataSourceSecretStores models data into format suitable for saving to
// Terraform state.
func flattenDataSourceSecretStores(remoteState []gofastly.SecretStore) []map[string]any {
	result := make([]map[string]any, len(remoteState))
	if len(remoteState) == 0 {
		return result
	}

	for i, resource := range remoteState {
		result[i] = map[string]any{
			"id":   resource.ID,
			"name": resource.Name,
		}
	}

	return result
}
