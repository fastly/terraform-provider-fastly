package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertMicrosoftTeamsIntegrations "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
)

func dataSourceFastlyNGWAFAlertMicrosoftTeamsIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead,
		Schema: map[string]*schema.Schema{
			"microsoft_teams_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Microsoft Teams alerts for a workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "The ID of the workspace alert.",
							Type:        schema.TypeString,
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The ID of the workspace.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Microsoft Teams alerts from workspace %s", workspaceID)

	remoteState, err := AlertMicrosoftTeamsIntegrations.List(ctx, conn, &AlertMicrosoftTeamsIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Microsoft Teams alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertMicrosoftTeamsIntegrations.Alert to []*AlertMicrosoftTeamsIntegrations.Alerts
	var AlertMicrosoftTeamsIntegrationsPtrs []*AlertMicrosoftTeamsIntegrations.Alert
	for i := range remoteState.Data {
		AlertMicrosoftTeamsIntegrationsPtrs = append(AlertMicrosoftTeamsIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("microsoft_teams_alerts", flattenNGWAFAlertMicrosoftTeamsIntegration(AlertMicrosoftTeamsIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Microsoft Teams alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertMicrosoftTeamsIntegration(remoteState []*AlertMicrosoftTeamsIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
