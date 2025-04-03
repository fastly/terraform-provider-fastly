package fastly

import (
	"context"
	"log"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceFastlyIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyIntegrationCreate,
		ReadContext:   resourceFastlyIntegrationRead,
		UpdateContext: resourceFastlyIntegrationUpdate,
		DeleteContext: resourceFastlyIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"config": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Configuration specific to the integration `type` (see documentation examples).",
				Elem:        schema.TypeString,
				Sensitive:   true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User submitted description of the integration.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User submitted name of the integration.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of the integration. One of: `mailinglist`, `microsoftteams`, `newrelic`, `pagerduty`, `slack`, `webhook`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"mailinglist", "microsoftteams", "newrelic", "pagerduty", "slack", "webhook"},
					false,
				)),
			},
		},
	}
}

func resourceFastlyIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.CreateIntegrationInput{
		Config: castToMapString(d.Get("config").(map[string]any)),
		Name:   gofastly.ToPointer(d.Get("name").(string)),
		Type:   gofastly.ToPointer(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	i, err := conn.CreateIntegration(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*i.ID)

	return resourceFastlyIntegrationRead(ctx, d, meta)
}

func resourceFastlyIntegrationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Integration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	i, err := conn.GetIntegration(&gofastly.GetIntegrationInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if i.Config != nil {
		err = d.Set("config", i.Config)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if i.Description != nil {
		err = d.Set("description", i.Description)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = d.Set("name", i.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("type", i.Type)
	if err != nil {
		return diag.FromErr(err)
	}

	if i.Type != nil && *i.Type == "mailinglist" && i.Status != nil && *i.Status != "confirmed" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Mailing list integration needs confirmation.",
				Detail:   "Please visit https://manage.fastly.com/observability/alerts/integrations to send a confirmation email and/or verify status.",
			},
		}
	}

	return nil
}

func resourceFastlyIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.UpdateIntegrationInput{
		Config: castToMapString(d.Get("config").(map[string]any)),
		ID:     d.Id(),
		Name:   gofastly.ToPointer(d.Get("name").(string)),
		Type:   gofastly.ToPointer(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	err := conn.UpdateIntegration(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyIntegrationRead(ctx, d, meta)
}

func resourceFastlyIntegrationDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteIntegration(&gofastly.DeleteIntegrationInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func castToMapString(m map[string]interface{}) map[string]string {
	result := map[string]string{}
	for k := range m {
		if v, ok := m[k].(string); ok {
			result[k] = v
		}
	}
	return result
}
