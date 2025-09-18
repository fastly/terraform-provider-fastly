package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	microsoftTeamsAlerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
)

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"webhook": {
				Description: "The Microsoft Teams webhook URL.",
				Required:    true,
				Type:        schema.TypeString,
				Sensitive:   !DisplaySensitiveFields,
			},
			"workspace_id": {
				Description: "The ID of the workspace.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := microsoftTeamsAlerts.CreateInput{
		Config: &microsoftTeamsAlerts.CreateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Microsoft Teams alert input: %#v", i)

	alert, err := microsoftTeamsAlerts.Create(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Microsoft Teams alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := microsoftTeamsAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Microsoft Teams alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := microsoftTeamsAlerts.Get(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Microsoft Teams alert not found '%s'", d.Id())
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

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := microsoftTeamsAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &microsoftTeamsAlerts.UpdateConfig{
			Webhook: gofastly.ToPointer(d.Get("webhook").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Microsoft Teams alert input: %#v", i)

	_, err := microsoftTeamsAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := microsoftTeamsAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Microsoft	Teams alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := microsoftTeamsAlerts.Delete(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertMicrosoftTeamsIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Microsoft Teams alert ID: %s", d.Id())

	workspaceID, alertMicrosoftTeamsIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertMicrosoftTeamsIntegrationID)

	return []*schema.ResourceData{d}, nil
}
