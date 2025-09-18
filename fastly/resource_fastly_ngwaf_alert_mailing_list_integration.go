package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	mailingListAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/mailinglist"
)

func resourceFastlyNGWAFAlertMailingListIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertMailingListIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertMailingListIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertMailingListIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertMailingListIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertMailingListIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"address": {
				Description: "Email address that the alert will use.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"workspace_id": {
				Description: "The ID of the workspace.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertMailingListIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := mailingListAlerts.CreateInput{
		Config: &mailingListAlerts.CreateConfig{
			Address: gofastly.ToPointer(d.Get("address").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF MailingList alert input: %#v", i)

	alert, err := mailingListAlerts.Create(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF MailingList alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertMailingListIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertMailingListIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := mailingListAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF MailingList alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := mailingListAlerts.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] MailingList alert not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("description", alert.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("address", alert.Config.Address); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertMailingListIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := mailingListAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &mailingListAlerts.UpdateConfig{
			Address: gofastly.ToPointer(d.Get("address").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF MailingList alert input: %#v", i)

	_, err := mailingListAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertMailingListIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertMailingListIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := mailingListAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF MailingList alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := mailingListAlerts.Delete(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertMailingListIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF MailingList alert ID: %s", d.Id())

	workspaceID, alertMailingListIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertMailingListIntegrationID)

	return []*schema.ResourceData{d}, nil
}
