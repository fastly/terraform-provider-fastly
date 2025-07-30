package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertDatadogIntegrations "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

func dataSourceFastlyNGWAFAlertDatadogIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertDatadogIntegrationRead,
		Schema: map[string]*schema.Schema{
			"datadog_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Datadog alerts for a workspace.",
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
				Description: "The id of the workspace that is being queried for Datadog alerts.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFAlertDatadogIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Datadog alerts from workspace %s", workspaceID)

	remoteState, err := AlertDatadogIntegrations.List(ctx, conn, &AlertDatadogIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Datadog alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertDatadogIntegrations.Alert to []*AlertDatadogIntegrations.Alerts
	var AlertDatadogIntegrationsPtrs []*AlertDatadogIntegrations.Alert
	for i := range remoteState.Data {
		AlertDatadogIntegrationsPtrs = append(AlertDatadogIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("datadog_alerts", flattenNGWAFAlertDatadogIntegration(AlertDatadogIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Datadog alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertDatadogIntegration(remoteState []*AlertDatadogIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
