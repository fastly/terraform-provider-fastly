package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

// go-fastly does not export type constants for these integration types, so
// they're defined here to sit alongside the ones it does export.
const (
	integrationTypeMailingList    = "mailinglist"
	integrationTypeMicrosoftTeams = "microsoftteams"
	integrationTypeNewRelic       = "newrelic"
	integrationTypePagerDuty      = "pagerduty"
	integrationTypeSlack          = "slack"
	integrationTypeWebhook        = "webhook"
)

// integrationTypes is the single source of truth for valid `type` values, so
// the schema's Description and ValidateDiagFunc can't drift out of sync with
// one another.
var integrationTypes = []string{
	gofastly.IntegrationTypeDatadog,
	gofastly.IntegrationTypeJiraIssue,
	gofastly.IntegrationTypeJSM,
	integrationTypeMailingList,
	integrationTypeMicrosoftTeams,
	integrationTypeNewRelic,
	gofastly.IntegrationTypeOpsGenie,
	integrationTypePagerDuty,
	integrationTypeSlack,
	gofastly.IntegrationTypeSplunkOnCall,
	integrationTypeWebhook,
}

// integrationTypeDescription renders integrationTypes into the schema
// Description text.
func integrationTypeDescription() string {
	quoted := make([]string, len(integrationTypes))
	for i, t := range integrationTypes {
		quoted[i] = fmt.Sprintf("`%s`", t)
	}
	return fmt.Sprintf("Type of the integration. One of: %s.", strings.Join(quoted, ", "))
}

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
				Sensitive:   !DisplaySensitiveFields,
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
				Type:             schema.TypeString,
				Required:         true,
				Description:      integrationTypeDescription(),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(integrationTypes, false)),
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

	i, err := conn.CreateIntegration(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*i.ID)

	return resourceFastlyIntegrationRead(ctx, d, meta)
}

func resourceFastlyIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Integration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	i, err := conn.GetIntegration(ctx, &gofastly.GetIntegrationInput{
		ID: d.Id(),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Integration not found (%s); removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	if i.Config != nil {
		// The API treats `config` as write-only for secret fields (e.g. `apikey`,
		// `token`): it's either entirely absent from the response, or present
		// with those fields stripped out. Merge the returned fields over the
		// configured value instead of overwriting it, so write-only fields
		// that aren't echoed back aren't seen as configuration drift.
		merged := d.Get("config").(map[string]any)
		for k, v := range i.Config {
			merged[k] = v
		}

		err = d.Set("config", merged)
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

	if i.Type != nil && *i.Type == integrationTypeMailingList && i.Status != nil && *i.Status != "confirmed" {
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

	err := conn.UpdateIntegration(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyIntegrationRead(ctx, d, meta)
}

func resourceFastlyIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteIntegration(ctx, &gofastly.DeleteIntegrationInput{
		ID: d.Id(),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}

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
