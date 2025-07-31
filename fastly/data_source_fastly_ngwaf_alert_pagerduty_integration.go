package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertPagerDutyIntegrations "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
)

func dataSourceFastlyNGWAFAlertPagerDutyIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertPagerDutyIntegrationRead,
		Schema: map[string]*schema.Schema{
			"pagerduty_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all PagerDuty alerts for a workspace.",
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
				Description: "The id of the workspace that is being queried for PagerDuty alerts.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFAlertPagerDutyIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF PagerDuty alerts from workspace %s", workspaceID)

	remoteState, err := AlertPagerDutyIntegrations.List(ctx, conn, &AlertPagerDutyIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching PagerDuty alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertPagerDutyIntegrations.Alert to []*AlertPagerDutyIntegrations.Alerts.
	var AlertPagerDutyIntegrationsPtrs []*AlertPagerDutyIntegrations.Alert
	for i := range remoteState.Data {
		AlertPagerDutyIntegrationsPtrs = append(AlertPagerDutyIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("pagerduty_alerts", flattenNGWAFAlertPagerDutyIntegration(AlertPagerDutyIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting PagerDuty alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertPagerDutyIntegration(remoteState []*AlertPagerDutyIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
