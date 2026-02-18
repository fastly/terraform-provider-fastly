package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertJiraIntegrations "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/jira"
)

func dataSourceFastlyNGWAFAlertJiraIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertJiraIntegrationRead,
		Schema: map[string]*schema.Schema{
			"jira_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Jira alerts for a workspace.",
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

func dataSourceFastlyNGWAFAlertJiraIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Jira alerts from workspace %s", workspaceID)

	remoteState, err := AlertJiraIntegrations.List(ctx, conn, &AlertJiraIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Jira alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertJiraIntegrations.Alert to []*AlertJiraIntegrations.Alerts
	var AlertJiraIntegrationsPtrs []*AlertJiraIntegrations.Alert
	for i := range remoteState.Data {
		AlertJiraIntegrationsPtrs = append(AlertJiraIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("jira_alerts", flattenNGWAFAlertJiraIntegration(AlertJiraIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Jira alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertJiraIntegration(remoteState []*AlertJiraIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
