package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

func dataSourceFastlyAIRuntimeControlProviderConnections() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAIRuntimeControlProviderConnectionsRead,
		Schema: map[string]*schema.Schema{
			"provider_connections": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The provider connections configured for the customer.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The base URL for the provider's API.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the provider connection was created.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the provider connection.",
						},
						"models": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The allowed AI model identifiers.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The human-readable name of the provider.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the provider connection was last updated.",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of provider connections returned.",
			},
		},
	}
}

func dataSourceFastlyAIRuntimeControlProviderConnectionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading AI Runtime Control provider connections")

	remoteState, err := providerconnection.List(ctx, conn, &providerconnection.ListInput{})
	if err != nil {
		return diag.Errorf("error fetching AI Runtime Control provider connections: %s", err)
	}

	all := remoteState.Data

	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("provider_connections", flattenAIRuntimeControlProviderConnections(all)); err != nil {
		return diag.Errorf("error setting provider_connections: %s", err)
	}
	if err := d.Set("total", len(all)); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenAIRuntimeControlProviderConnections(remoteState []providerconnection.ProviderConnection) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, pc := range remoteState {
		result[i] = map[string]any{
			"id":         pc.ID,
			"name":       pc.Name,
			"models":     pc.Models,
			"base_url":   pc.BaseURL,
			"created_at": pc.CreatedAt,
			"updated_at": pc.UpdatedAt,
		}
	}

	return result
}
