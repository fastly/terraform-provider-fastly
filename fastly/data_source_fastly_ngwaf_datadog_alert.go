package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	datadogAlerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

func dataSourceFastlyNGWAFDatadogAlert() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFDatadogAlertRead,
		Schema: map[string]*schema.Schema{
			"datadog_alerts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all datadog alerts for a workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Description: "User-submitted description of the integration",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"id": {
							Computed:    true,
							Description: "Base62-encoded representation of a UUID used to uniquely identify the redaction",
							Type:        schema.TypeString,
						},
						"integration_key": {
							Description: "The Datadog integration key.",
							Required:    true,
							Type:        schema.TypeString,
						},
						"integration_site": {
							Description: "The Datadog site.",
							Required:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The id of the workspace that is being queried for datadog alerts.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFDatadogAlertRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF datadog alerts from workspace %s", workspaceID)

	remoteState, err := datadogAlerts.List(ctx, conn, &datadogAlerts.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching datadog alerts: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []datadogAlerts.Alert to []*datadogAlerts.Alerts
	var datadogAlertsPtrs []*datadogAlerts.Alert
	for i := range remoteState.Data {
		datadogAlertsPtrs = append(datadogAlertsPtrs, &remoteState.Data[i])
	}

	if err := d.Set("datadog_alerts", flattenNGWAFDatadogAlert(datadogAlertsPtrs)); err != nil {
		return diag.Errorf("error setting datadog alerts: %s", err)
	}

	return nil
}

func flattenNGWAFDatadogAlert(remoteState []*datadogAlerts.Alert) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id":               r.ID,
			"description":      r.Description,
			"integration_key":  r.Config.Key,
			"integration_site": r.Config.Site,
		}
	}

	return result
}
