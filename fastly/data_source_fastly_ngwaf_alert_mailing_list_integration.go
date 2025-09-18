package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertMailingListIntegrations "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/mailinglist"
)

func dataSourceFastlyNGWAFAlertMailingListIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertMailingListIntegrationRead,
		Schema: map[string]*schema.Schema{
			"mailing_list_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Mailing List alerts for a workspace.",
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

func dataSourceFastlyNGWAFAlertMailingListIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Mailing List alerts from workspace %s", workspaceID)

	remoteState, err := AlertMailingListIntegrations.List(ctx, conn, &AlertMailingListIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Mailing List alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertMailingListIntegrations.Alert to []*AlertMailingListIntegrations.Alerts
	var AlertMailingListIntegrationsPtrs []*AlertMailingListIntegrations.Alert
	for i := range remoteState.Data {
		AlertMailingListIntegrationsPtrs = append(AlertMailingListIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("mailing_list_alerts", flattenNGWAFAlertMailingListIntegration(AlertMailingListIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Mailing List alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertMailingListIntegration(remoteState []*AlertMailingListIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
