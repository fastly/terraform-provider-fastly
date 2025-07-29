package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	wsr "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/thresholds"
)

func dataSourceFastlyNGWAFThresholds() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFThresholdsRead,
		Schema: map[string]*schema.Schema{
			"thresholds": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all thresholds for a given workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action to take when threshold is exceeded.",
						},
						"dont_notify": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to silence notifications when action is taken.",
						},
						"duration": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Duration the action is in place.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this threshold is active.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the threshold.",
						},
						"interval": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Threshold interval in seconds.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Threshold limit.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Threshold name.",
						},
						"signal": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Signal this threshold is acting on. For custom Signals, append signal with `site.<name>`. For System Signals, input the <name> field only.",
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

func dataSourceFastlyNGWAFThresholdsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)
	log.Printf("[DEBUG] Reading NGWAF thresholds for workspace: %s", workspaceID)

	remoteState, err := wsr.List(ctx, conn, &wsr.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching thresholds: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var thresholdPtrs []*wsr.Threshold
	for i := range remoteState.Data {
		thresholdPtrs = append(thresholdPtrs, &remoteState.Data[i])
	}

	if err := d.Set("thresholds", flattenNGWAFThresholds(thresholdPtrs)); err != nil {
		return diag.Errorf("error setting thresholds: %s", err)
	}

	return nil
}

func flattenNGWAFThresholds(remoteState []*wsr.Threshold) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, t := range remoteState {
		result[i] = map[string]any{
			"id":          t.ThresholdID,
			"action":      t.Action,
			"dont_notify": t.DontNotify,
			"duration":    t.Duration,
			"enabled":     t.Enabled,
			"interval":    t.Interval,
			"limit":       t.Limit,
			"name":        t.Name,
			"signal":      t.Signal,
		}
	}

	return result
}
