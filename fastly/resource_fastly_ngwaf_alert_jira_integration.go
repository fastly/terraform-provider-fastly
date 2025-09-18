package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	jiraAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/jira"
)

func resourceFastlyNGWAFAlertJiraIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFAlertJiraIntegrationCreate,
		ReadContext:   resourceFastlyNGWAFAlertJiraIntegrationRead,
		UpdateContext: resourceFastlyNGWAFAlertJiraIntegrationUpdate,
		DeleteContext: resourceFastlyNGWAFAlertJiraIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFAlertJiraIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description of the alert.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"host": {
				Description: "The name of the Jira instance.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"issue_type": {
				Description: "The Jira issue type associated with the ticket.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"key": {
				Description: "The Jira key.",
				Required:    true,
				Type:        schema.TypeString,
				Sensitive:   !DisplaySensitiveFields,
			},
			"project": {
				Description: "The Jira project where the issue will be created.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"username": {
				Description: "The Jira username of the user who created the ticket.",
				Required:    true,
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

func resourceFastlyNGWAFAlertJiraIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := jiraAlerts.CreateInput{
		Config: &jiraAlerts.CreateConfig{
			Host:      gofastly.ToPointer(d.Get("host").(string)),
			IssueType: gofastly.ToPointer(d.Get("issue_type").(string)),
			Key:       gofastly.ToPointer(d.Get("key").(string)),
			Project:   gofastly.ToPointer(d.Get("project").(string)),
			Username:  gofastly.ToPointer(d.Get("username").(string)),
		},
		Description: gofastly.ToPointer(d.Get("description").(string)),
		Events:      gofastly.ToPointer([]string{"flag"}),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF Jira alert input: %#v", i)

	alert, err := jiraAlerts.Create(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF Jira alert result: %#v", alert)

	d.SetId(alert.ID)

	return resourceFastlyNGWAFAlertJiraIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertJiraIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := jiraAlerts.GetInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Jira alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	alert, err := jiraAlerts.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Jira alert not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("description", alert.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("host", alert.Config.Host); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("issue_type", alert.Config.IssueType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", alert.Config.Key); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", alert.Config.Project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", alert.Config.Username); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertJiraIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := jiraAlerts.UpdateInput{
		AlertID: gofastly.ToPointer(d.Id()),
		Config: &jiraAlerts.UpdateConfig{
			Host:      gofastly.ToPointer(d.Get("host").(string)),
			IssueType: gofastly.ToPointer(d.Get("issue_type").(string)),
			Key:       gofastly.ToPointer(d.Get("key").(string)),
			Project:   gofastly.ToPointer(d.Get("project").(string)),
			Username:  gofastly.ToPointer(d.Get("username").(string)),
		},
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Jira alert input: %#v", i)

	_, err := jiraAlerts.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFAlertJiraIntegrationRead(ctx, d, meta)
}

func resourceFastlyNGWAFAlertJiraIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := jiraAlerts.DeleteInput{
		AlertID:     gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Jira alert input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := jiraAlerts.Delete(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFAlertJiraIntegrationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF Jira alert ID: %s", d.Id())

	workspaceID, alertJiraIntegrationID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<alertID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(alertJiraIntegrationID)

	return []*schema.ResourceData{d}, nil
}
