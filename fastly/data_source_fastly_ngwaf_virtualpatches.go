package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces"
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

	log.Printf("[DEBUG] Reading NGWAF workspaces")

	remoteState, err := ws.List(ctx, conn, &ws.{})
	if err != nil {
		return diag.Errorf("error fetching workspaces: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []workspaces.Workspace to []*workspaces.Workspace
	var workspacePtrs []*ws.Workspace
	for i := range remoteState.Data {
		workspacePtrs = append(workspacePtrs, &remoteState.Data[i])
	}

	if err := d.Set("workspaces", flattenNGWAFVirtualPatches(workspacePtrs)); err != nil {
		return diag.Errorf("error setting workspaces: %s", err)
	}

	return nil
}

func flattenNGWAFVirtualPatches(remoteState []*ws.Workspace) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, w := range remoteState {
		result[i] = map[string]any{
			"id":   w.WorkspaceID,
			"name": w.Name,
		}
	}

	return result
}
