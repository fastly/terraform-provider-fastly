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
			"virtualpatches": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all virtual patches for a given workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workspace_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Base62-encoded representation of a UUID used to uniquely identify the workspace",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyNGWAFVirtualPatchesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading NGWAF virtual patches")

	remoteState, err := ws.List(ctx, conn, &ws.ListInput{})
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

	if err := d.Set("virtualpatches", flattenNGWAFVirtualPatches(virtualpatchPtrs)); err != nil {
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
