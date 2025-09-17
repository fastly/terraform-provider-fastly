package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	webhookAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/webhook"
)

func resourceFastlyNGWAFAlertWebhookIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertWebhookIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertWebhookIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertWebhookIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertWebhookIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertWebhookIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"webhook": {
				Description: "The webhook URL.",
				Required:    true,
				Type:        schema.TypeString,
				Sensitive:   true,
			},
			"workspace_id": {
				Description: "The ID of the workspace.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertWebhookIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := webhookAlerts.CreateInput{
		Config: &webhookAlerts.CreateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Webhook alert input: %#v", i)

	alert, err := webhookAlerts.Create(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Webhook alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertWebhookIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertWebhookIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := webhookAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Webhook alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := webhookAlerts.Get(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Webhook alert not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("description", alert.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("webhook", alert.Config.Webhook); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertWebhookIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := webhookAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &webhookAlerts.UpdateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Webhook alert input: %#v", i)

	_, err := webhookAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertWebhookIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertWebhookIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := webhookAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Webhook alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := webhookAlerts.Delete(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertWebhookIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Webhook alert ID: %s", d.Id())

	workspaceID, alertWebhookIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertWebhookIntegrationID)

	return []*schema.ResourceData{d}, nil
}
