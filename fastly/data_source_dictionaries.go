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

func dataSourceFastlyDictionaries() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyDictionariesRead,
		Schema: map[string]*schema.Schema{
			"dictionaries": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all dictionaries for the version of the service.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the Dictionary.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name for the Dictionary.",
						},
						"write_only": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if items in the dictionary are readable or not.",
						},
					},
				},
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Alphanumeric string identifying the service.",
			},
			"service_version": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Integer identifying a service version.",
			},
		},
	}
}

func dataSourceFastlyDictionariesRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading dictionaries")

	remoteState, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
		ServiceID:      d.Get("service_id").(string),
		ServiceVersion: d.Get("service_version").(int),
	})
	if err != nil {
		return diag.Errorf("error fetching dictionaries: %s", err)
	}

	hashBase, _ := json.Marshal(remoteState)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("dictionaries", flattenDataSourceDictionaries(remoteState)); err != nil {
		return diag.Errorf("error setting dictionaries: %s", err)
	}

	return nil
}

// flattenDataSourceDictionaries models data into format suitable for saving to
// Terraform state.
func flattenDataSourceDictionaries(remoteState []*gofastly.Dictionary) []map[string]any {
	result := make([]map[string]any, len(remoteState))
	if len(remoteState) == 0 {
		return result
	}

	for i, resource := range remoteState {
		result[i] = map[string]any{
			"id":         resource.ID,
			"name":       resource.Name,
			"write_only": resource.WriteOnly,
		}
	}

	return result
}
