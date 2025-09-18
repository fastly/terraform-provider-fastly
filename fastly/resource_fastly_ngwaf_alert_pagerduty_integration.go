package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	pagerdutyAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
)

func resourceFastlyNGWAFAlertPagerDutyIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertPagerDutyIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertPagerDutyIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertPagerDutyIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertPagerDutyIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertPagerDutyIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"key": {
				Description:  "The PagerDuty integration key.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(32, 32),
				Sensitive:    !DisplaySensitiveFields,
			},
			"workspace_id": {
				Description: "The ID of the workspace.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertPagerDutyIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := pagerdutyAlerts.CreateInput{
		Config: &pagerdutyAlerts.CreateConfig{
			Key: gofastly.ToPointer(d.Get("key").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF PagerDuty alert input: %#v", i)

	alert, err := pagerdutyAlerts.Create(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF PagerDuty alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertPagerDutyIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertPagerDutyIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := pagerdutyAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF PagerDuty alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := pagerdutyAlerts.Get(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] PagerDuty alert not found '%s'", d.Id())
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

func resourceFastlyNGWAFAlertPagerDutyIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := pagerdutyAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &pagerdutyAlerts.UpdateConfig{
			Key: gofastly.ToPointer(d.Get("key").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF PagerDuty alert input: %#v", i)

	_, err := pagerdutyAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertPagerDutyIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertPagerDutyIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := pagerdutyAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF PagerDuty alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := pagerdutyAlerts.Delete(gofastly.NewContextForResourceID(ctx, workspaceID), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertPagerDutyIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF PagerDuty alert ID: %s", d.Id())

	workspaceID, alertPagerDutyIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertPagerDutyIntegrationID)

	return []*schema.ResourceData{d}, nil
}
