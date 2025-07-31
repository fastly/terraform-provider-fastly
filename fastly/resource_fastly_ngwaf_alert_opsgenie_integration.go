package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	opsgenieAlerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/opsgenie"
)

func resourceFastlyNGWAFAlertOpsgenieIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertOpsgenieIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertOpsgenieIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertOpsgenieIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertOpsgenieIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertOpsgenieIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "User-submitted description of the alert",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"key": {
				Description: "The Opsgenie key.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"workspace_id": {
				Description: "The id of the workspace this alert belongs to.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertOpsgenieIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := opsgenieAlerts.CreateInput{
		Config: &opsgenieAlerts.CreateConfig{
			Key: gofastly.ToPointer(d.Get("key").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Opsgenie alert input: %#v", i)

	alert, err := opsgenieAlerts.Create(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Opsgenie alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertOpsgenieIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertOpsgenieIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := opsgenieAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Opsgenie alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := opsgenieAlerts.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Opsgenie alert not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("description", alert.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", alert.Config.Key); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertOpsgenieIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := opsgenieAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &opsgenieAlerts.UpdateConfig{
			Key: gofastly.ToPointer(d.Get("key").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Opsgenie alert input: %#v", i)

	_, err := opsgenieAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertOpsgenieIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertOpsgenieIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := opsgenieAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Opsgenie alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := opsgenieAlerts.Delete(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertOpsgenieIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Opsgenie alert ID: %s", d.Id())

	workspaceID, alertOpsgenieIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertOpsgenieIntegrationID)

	return []*schema.ResourceData{d}, nil
}
