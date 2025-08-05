package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/virtualpatches"
)

func dataSourceFastlyNGWAFVirtualPatches() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFVirtualPatchesRead,
		Schema: map[string]*schema.Schema{
			"virtual_patches": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all virtual patches for a given workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the virtual patch is enabled or disabled.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the virtual patch.",
						},
						"mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action to take when a signal for the virtual patch is detected. One of `log` or `block`.",
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace.",
			},
		},
	}
}

func dataSourceFastlyNGWAFVirtualPatchesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)
	log.Printf("[DEBUG] Reading NGWAF virtual patches for workspace: %s", workspaceID)

	remoteState, err := ws.List(ctx, conn, &ws.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching virtual patches: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var virtualpatchPtrs []*ws.VirtualPatch
	for i := range remoteState.Data {
		virtualpatchPtrs = append(virtualpatchPtrs, &remoteState.Data[i])
	}

	if err := d.Set("virtual_patches", flattenNGWAFVirtualPatches(virtualpatchPtrs)); err != nil {
		return diag.Errorf("error setting virtualpatches: %s", err)
	}

	return nil
}

func flattenNGWAFVirtualPatches(remoteState []*ws.VirtualPatch) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, w := range remoteState {
		result[i] = map[string]any{
			"id":      w.ID,
			"enabled": w.Enabled,
			"mode":    w.Mode,
		}
	}

	return result
}
