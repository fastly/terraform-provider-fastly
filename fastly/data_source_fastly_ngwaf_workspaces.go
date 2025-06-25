package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	ws "github.com/fastly/go-fastly/v10/fastly/ngwaf/v1/workspaces"
)

func dataSourceFastlyNGWAFWorkspaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFWorkspacesRead,
		Schema: map[string]*schema.Schema{
			"workspaces": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all NGWAF Workspaces.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the NGWAF Workspace.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the NGWAF Workspace.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyNGWAFWorkspacesRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading NGWAF Workspaces")

	remoteState, err := ws.List(conn, &ws.ListInput{})
	if err != nil {
		return diag.Errorf("error fetching NGWAF Workspaces: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []workspaces.Workspace to []*workspaces.Workspace
	var workspacePtrs []*ws.Workspace
	for i := range remoteState.Data {
		workspacePtrs = append(workspacePtrs, &remoteState.Data[i])
	}

	if err := d.Set("workspaces", flattenNGWAFWorkspaces(workspacePtrs)); err != nil {
		return diag.Errorf("error setting workspaces: %s", err)
	}

	return nil
}

func flattenNGWAFWorkspaces(remoteState []*ws.Workspace) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, w := range remoteState {
		result[i] = map[string]any{
			"id":   w.WorkspaceID,
			"name": w.Name,
		}
	}

	return result
}
