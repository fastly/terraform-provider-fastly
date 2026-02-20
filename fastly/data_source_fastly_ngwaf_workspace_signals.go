package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/signals"
)

func dataSourceFastlyNGWAFWorkspaceSignals() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFWorkspaceSignalsRead,
		Schema: map[string]*schema.Schema{
			"signals": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of custom signals.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the signal.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the signal.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the signal.",
						},
						"tag_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The tag name of the signal.",
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

func dataSourceFastlyNGWAFWorkspaceSignalsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF workspace signals for workspace: %s", workspaceID)

	scopeObj := &scope.Scope{
		Type:      scope.ScopeTypeWorkspace,
		AppliesTo: []string{workspaceID},
	}

	remoteState, err := signals.List(ctx, conn, &signals.ListInput{
		Scope: scopeObj,
	})
	if err != nil {
		return diag.Errorf("error fetching signals: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var signalPtrs []*signals.Signal
	for i := range remoteState.Data {
		signalPtrs = append(signalPtrs, &remoteState.Data[i])
	}

	if err := d.Set("signals", flattenNGWAFWorkspaceSignals(signalPtrs)); err != nil {
		return diag.Errorf("error setting signals: %s", err)
	}

	return nil
}

func flattenNGWAFWorkspaceSignals(remoteState []*signals.Signal) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, signal := range remoteState {
		result[i] = map[string]any{
			"id":          signal.SignalID,
			"name":        signal.Name,
			"tag_name":    signal.ReferenceID,
			"description": signal.Description,
		}
	}

	return result
}
