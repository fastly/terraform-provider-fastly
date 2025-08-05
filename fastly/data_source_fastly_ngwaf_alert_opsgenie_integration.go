package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertOpsgenieIntegrations "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/opsgenie"
)

func dataSourceFastlyNGWAFAlertOpsgenieIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertOpsgenieIntegrationRead,
		Schema: map[string]*schema.Schema{
			"opsgenie_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Opsgenie alerts for a workspace.",
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

func dataSourceFastlyNGWAFAlertOpsgenieIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF Opsgenie alerts from workspace %s", workspaceID)

	remoteState, err := AlertOpsgenieIntegrations.List(ctx, conn, &AlertOpsgenieIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching Opsgenie alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertOpsgenieIntegrations.Alert to []*AlertOpsgenieIntegrations.Alerts
	var AlertOpsgenieIntegrationsPtrs []*AlertOpsgenieIntegrations.Alert
	for i := range remoteState.Data {
		AlertOpsgenieIntegrationsPtrs = append(AlertOpsgenieIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("opsgenie_alerts", flattenNGWAFAlertOpsgenieIntegration(AlertOpsgenieIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting Opsgenie alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertOpsgenieIntegration(remoteState []*AlertOpsgenieIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
