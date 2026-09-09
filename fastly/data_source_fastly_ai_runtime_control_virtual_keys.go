package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
)

func dataSourceFastlyAIRuntimeControlVirtualKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAIRuntimeControlVirtualKeysRead,
		Schema: map[string]*schema.Schema{
			"include_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Include deleted virtual keys in the results. Default `false`.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by AI model identifier.",
			},
			// The Terraform SDK reserves the field name "provider" on every
			// data source, so the API's `provider` filter is exposed as
			// `provider_name`.
			"provider_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by AI model provider name.",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by a substring match on the virtual key name.",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of virtual keys returned.",
			},
			"virtual_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The virtual keys matching the given filters. Access tokens are never returned by this endpoint.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the virtual key was created.",
						},
						"created_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display name of the user who created the virtual key.",
						},
						"customer_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the customer that owns the virtual key.",
						},
						"deleted_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the virtual key was deleted, if applicable.",
						},
						"expires_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration timestamp of the virtual key, if any.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the virtual key.",
						},
						"last_used_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the virtual key was last used, if applicable.",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The AI model identifier.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The human-readable name of the virtual key.",
						},
						"provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The AI model provider name.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the virtual key.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp (UTC) of when the virtual key was last updated.",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the user who created the virtual key.",
						},
						"user_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display name of the user who created the virtual key.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyAIRuntimeControlVirtualKeysRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading AI Runtime Control virtual keys")

	var all []key.VirtualKeyListItem
	// Request the largest page the API allows to minimise round trips.
	limit := 100

	i := key.ListInput{
		Limit:          &limit,
		IncludeDeleted: gofastly.ToPointer(d.Get("include_deleted").(bool)),
	}
	if v, ok := d.GetOk("model"); ok {
		i.Model = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("provider_name"); ok {
		i.Provider = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("search"); ok {
		i.Search = gofastly.ToPointer(v.(string))
	}

	for {
		remoteState, err := key.List(ctx, conn, &i)
		if err != nil {
			return diag.Errorf("error fetching AI Runtime Control virtual keys: %s", err)
		}

		all = append(all, remoteState.Data...)

		if remoteState.Meta.NextCursor == "" {
			break
		}
		i.Cursor = &remoteState.Meta.NextCursor
	}

	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("virtual_keys", flattenAIRuntimeControlVirtualKeys(all)); err != nil {
		return diag.Errorf("error setting virtual_keys: %s", err)
	}
	if err := d.Set("total", len(all)); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenAIRuntimeControlVirtualKeys(remoteState []key.VirtualKeyListItem) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, vk := range remoteState {
		result[i] = map[string]any{
			"id":           vk.ID,
			"type":         vk.Type,
			"name":         vk.Name,
			"model":        vk.Model,
			"provider":     vk.Provider,
			"user_id":      vk.UserID,
			"user_name":    vk.UserName,
			"customer_id":  vk.CustomerID,
			"created_at":   vk.CreatedAt,
			"created_by":   vk.CreatedBy,
			"updated_at":   vk.UpdatedAt,
			"expires_at":   gofastly.ToValue(vk.ExpiresAt),
			"deleted_at":   gofastly.ToValue(vk.DeletedAt),
			"last_used_at": gofastly.ToValue(vk.LastUsedAt),
		}
	}

	return result
}
