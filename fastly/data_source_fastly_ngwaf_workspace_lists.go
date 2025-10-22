package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

func dataSourceFastlyNGWAFWorkspaceLists() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFWorkspaceListsRead,
		Schema: map[string]*schema.Schema{
			"lists": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of lists.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time in ISO 8601 format when the list was created.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the list.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the list.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the list.",
						},
						"reference_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reference ID of the list.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the list.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time in ISO 8601 format when the list was last updated.",
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

func dataSourceFastlyNGWAFWorkspaceListsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF workspace lists for workspace: %s", workspaceID)

	scopeObj := &scope.Scope{
		Type:      scope.ScopeTypeWorkspace,
		AppliesTo: []string{workspaceID},
	}

	remoteState, err := lists.ListLists(ctx, conn, &lists.ListInput{
		Scope: scopeObj,
	})
	if err != nil {
		return diag.Errorf("error fetching lists: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var listPtrs []*lists.List
	for i := range remoteState.Data {
		listPtrs = append(listPtrs, &remoteState.Data[i])
	}

	if err := d.Set("lists", flattenNGWAFWorkspaceLists(listPtrs)); err != nil {
		return diag.Errorf("error setting lists: %s", err)
	}

	return nil
}

func flattenNGWAFWorkspaceLists(remoteState []*lists.List) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, list := range remoteState {
		result[i] = map[string]any{
			"created_at":   list.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"description":  list.Description,
			"id":           list.ListID,
			"name":         list.Name,
			"reference_id": list.ReferenceID,
			"type":         list.Type,
			"updated_at":   list.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result
}
