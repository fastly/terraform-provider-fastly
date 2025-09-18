package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	Threshold "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/thresholds"
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
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the threshold.",
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

	remoteState, err := Threshold.List(ctx, conn, &Threshold.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching thresholds: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var thresholdPtrs []*Threshold.Threshold
	for i := range remoteState.Data {
		thresholdPtrs = append(thresholdPtrs, &remoteState.Data[i])
	}

	if err := d.Set("thresholds", flattenNGWAFThresholds(thresholdPtrs)); err != nil {
		return diag.Errorf("error setting thresholds: %s", err)
	}

	return nil
}

func flattenNGWAFThresholds(remoteState []*Threshold.Threshold) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, t := range remoteState {
		result[i] = map[string]any{
			"id": t.ThresholdID,
		}
	}

	return result
}
