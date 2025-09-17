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
	ddalerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

func resourceFastlyNGWAFAlertDatadogIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertDatadogIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertDatadogIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertDatadogIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertDatadogIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertDatadogIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"key": {
				Description:  "The Datadog key.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(9, 9),
				Sensitive:    true,
			},
			"site": {
				Description:  "The Datadog site.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(3, 3),
			},
			"workspace_id": {
				Description: "The ID of the workspace.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceFastlyNGWAFAlertDatadogIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := ddalerts.CreateInput{
		Config: &ddalerts.CreateConfig{
			Key:  gofastly.ToPointer(d.Get("key").(string)),
			Site: gofastly.ToPointer(d.Get("site").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Datadog alert input: %#v", i)

	alert, err := ddalerts.Create(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Datadog alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertDatadogIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertDatadogIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := ddalerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Datadog alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := ddalerts.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Datadog alert not found '%s'", d.Id())
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
	if err := d.Set("site", alert.Config.Site); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertDatadogIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ddalerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &ddalerts.UpdateConfig{
			Key:  gofastly.ToPointer(d.Get("key").(string)),
			Site: gofastly.ToPointer(d.Get("site").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Datadog alert input: %#v", i)

	_, err := ddalerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertDatadogIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertDatadogIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := ddalerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Datadog alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := ddalerts.Delete(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertDatadogIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Datadog alert ID: %s", d.Id())

	workspaceID, alertDatadogIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertDatadogIntegrationID)

	return []*schema.ResourceData{d}, nil
}
