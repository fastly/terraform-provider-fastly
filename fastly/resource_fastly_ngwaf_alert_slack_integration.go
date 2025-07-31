package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	slackAlerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/slack"
)

func resourceFastlyNGWAFAlertSlackIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertSlackIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertSlackIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertSlackIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertSlackIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertSlackIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "User-submitted description of the alert",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"webhook": {
				Description: "The Slack webhook URL.",
				Required:    true,
				Type:        schema.TypeString,
				Sensitive:   true,
			},
			"workspace_id": {
				Description: "The id of the workspace this alert belongs to.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertSlackIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := slackAlerts.CreateInput{
		Config: &slackAlerts.CreateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Slack alert input: %#v", i)

	alert, err := slackAlerts.Create(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Slack alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertSlackIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertSlackIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := slackAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Slack alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := slackAlerts.Get(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Slack alert not found '%s'", d.Id())
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

func resourceFastlyNGWAFAlertSlackIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := slackAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &slackAlerts.UpdateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Slack alert input: %#v", i)

	_, err := slackAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertSlackIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertSlackIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := slackAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Slack alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := slackAlerts.Delete(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertSlackIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Slack alert ID: %s", d.Id())

	workspaceID, alertSlackIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertSlackIntegrationID)

	return []*schema.ResourceData{d}, nil
}
