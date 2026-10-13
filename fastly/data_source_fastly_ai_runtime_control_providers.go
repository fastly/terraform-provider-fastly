package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	arcprovider "github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/provider"
)

func dataSourceFastlyAIRuntimeControlProviders() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAIRuntimeControlProvidersRead,
		Schema: map[string]*schema.Schema{
			"providers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The AI providers supported by AI Runtime Control, with each provider's available models nested within.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_base_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The default API base URL for the provider.",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The human-readable provider name, e.g. `Anthropic`.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider identifier, e.g. `anthropic`.",
						},
						"models": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The models available for the provider.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The human-readable model name, e.g. `Claude Sonnet 4`.",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The model identifier, e.g. `claude-sonnet-4-20250514`.",
									},
									"provider_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the provider this model belongs to.",
									},
								},
							},
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of providers returned.",
			},
		},
	}
}

func dataSourceFastlyAIRuntimeControlProvidersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading AI Runtime Control providers")

	remoteState, err := arcprovider.List(ctx, conn)
	if err != nil {
		return diag.Errorf("error fetching AI Runtime Control providers: %s", err)
	}

	hashBase, _ := json.Marshal(remoteState)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("providers", flattenAIRuntimeControlProviders(remoteState.Data)); err != nil {
		return diag.Errorf("error setting providers: %s", err)
	}
	if err := d.Set("total", remoteState.Meta.Total); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenAIRuntimeControlProviders(remoteState []arcprovider.Provider) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, p := range remoteState {
		result[i] = map[string]any{
			"id":               p.ID,
			"display_name":     p.DisplayName,
			"default_base_url": p.DefaultBaseURL,
			"models":           flattenAIRuntimeControlModels(p.Models),
		}
	}

	return result
}

func flattenAIRuntimeControlModels(remoteState []arcprovider.Model) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, m := range remoteState {
		result[i] = map[string]any{
			"id":           m.ID,
			"display_name": m.DisplayName,
			"provider_id":  m.ProviderID,
		}
	}

	return result
}
