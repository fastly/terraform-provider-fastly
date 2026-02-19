package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	AlertWebhookIntegrations "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/webhook"
)

func dataSourceFastlyNGWAFAlertWebhookIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAlertWebhookIntegrationRead,
		Schema: map[string]*schema.Schema{
			"webhook_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all webhook alerts for a workspace.",
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

func dataSourceFastlyNGWAFAlertWebhookIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF webhook alerts from workspace %s", workspaceID)

	remoteState, err := AlertWebhookIntegrations.List(ctx, conn, &AlertWebhookIntegrations.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching webhook alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []AlertWebhookIntegrations.Alert to []*AlertWebhookIntegrations.Alerts
	var AlertWebhookIntegrationsPtrs []*AlertWebhookIntegrations.Alert
	for i := range remoteState.Data {
		AlertWebhookIntegrationsPtrs = append(AlertWebhookIntegrationsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("webhook_alerts", flattenNGWAFAlertWebhookIntegration(AlertWebhookIntegrationsPtrs)); err != nil {
		return diag.Errorf("error setting webhook alerts: %s", err)
	}

	return nil
}

func flattenNGWAFAlertWebhookIntegration(remoteState []*AlertWebhookIntegrations.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id": r.ID,
		}
	}

	return result
}
