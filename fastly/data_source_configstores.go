package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyConfigStores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyConfigStoresRead,
		Schema: map[string]*schema.Schema{
			"stores": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Config Stores.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the Config Store.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name for the Config Store.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyConfigStoresRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading Config Stores")

	remoteState, err := conn.ListConfigStores(&gofastly.ListConfigStoresInput{})
	if err != nil {
		return diag.Errorf("error fetching Config Stores: %s", err)
	}

	hashBase, _ := json.Marshal(remoteState)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("stores", flattenDataSourceConfigStores(remoteState)); err != nil {
		return diag.Errorf("error setting stores: %s", err)
	}

	return nil
}

// flattenDataSourceConfigStores models data into format suitable for saving to
// Terraform state.
func flattenDataSourceConfigStores(remoteState []*gofastly.ConfigStore) []map[string]any {
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
