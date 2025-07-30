package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertSlackIntegrations "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/slack"
)

func dataSourceFastlyNGWAFAlertSlackIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertSlackIntegrationRead,
		Schema: map[string]*schema.Schema{
			"slack_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Slack alerts for a workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "Base62-encoded representation of a UUID used to uniquely identify the alert",
							Type:        schema.TypeString,
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The id of the workspace that is being queried for Slack alerts.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFAlertSlackIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Slack alerts from workspace %s", workspaceID)

	remoteState, err := AlertSlackIntegrations.List(ctx, conn, &AlertSlackIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Slack alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertSlackIntegrations.Alert to []*AlertSlackIntegrations.Alerts
	var AlertSlackIntegrationsPtrs []*AlertSlackIntegrations.Alert
	for i := range remoteState.Data {
		AlertSlackIntegrationsPtrs = append(AlertSlackIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("slack_alerts", flattenNGWAFAlertSlackIntegration(AlertSlackIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Slack alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertSlackIntegration(remoteState []*AlertSlackIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
