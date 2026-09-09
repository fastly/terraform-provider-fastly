package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

func resourceFastlyAIRuntimeControlProviderConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAIRuntimeControlProviderConnectionCreate,
		ReadContext:   resourceFastlyAIRuntimeControlProviderConnectionRead,
		UpdateContext: resourceFastlyAIRuntimeControlProviderConnectionUpdate,
		DeleteContext: resourceFastlyAIRuntimeControlProviderConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   !DisplaySensitiveFields,
				Description: "The provider's secret key used for authentication. Never returned by the API, so it cannot be reconciled with the remote state.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The base URL for the provider's API, e.g. `https://api.openai.com`. Must not include a `/v1` suffix, which the API appends automatically. Trailing slashes are stripped.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (UTC) of when the provider connection was created.",
			},
			"models": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The set of allowed AI model identifiers, e.g. `[\"gpt-4o\", \"gpt-4o-mini\"]`.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A human-readable name for the provider. Changing this forces a new provider connection to be created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (UTC) of when the provider connection was last updated.",
			},
		},
	}
}

func resourceFastlyAIRuntimeControlProviderConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := providerconnection.CreateInput{
		Name:    gofastly.ToPointer(d.Get("name").(string)),
		BaseURL: gofastly.ToPointer(d.Get("base_url").(string)),
		APIKey:  gofastly.ToPointer(d.Get("api_key").(string)),
		Models:  expandAIRuntimeControlModels(d.Get("models").(*schema.Set)),
	}

	// api_key is deliberately omitted from this log line; do not log the full
	// input struct, as it contains the provider's secret key.
	log.Printf("[DEBUG] CREATE: AI Runtime Control provider connection input: name=%q base_url=%q models=%v", gofastly.ToValue(i.Name), gofastly.ToValue(i.BaseURL), i.Models)

	pc, err := providerconnection.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pc.ID)

	return resourceFastlyAIRuntimeControlProviderConnectionRead(ctx, d, meta)
}

func resourceFastlyAIRuntimeControlProviderConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := providerconnection.GetInput{
		ID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: AI Runtime Control provider connection input: %#v", i)

	pc, err := providerconnection.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] AI Runtime Control provider connection not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", pc.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("base_url", pc.BaseURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("models", pc.Models); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", pc.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", pc.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyAIRuntimeControlProviderConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// The API applies a partial update, so only send what actually changed.
	i := providerconnection.UpdateInput{
		ID: gofastly.ToPointer(d.Id()),
	}

	if d.HasChange("base_url") {
		i.BaseURL = gofastly.ToPointer(d.Get("base_url").(string))
	}
	if d.HasChange("api_key") {
		i.APIKey = gofastly.ToPointer(d.Get("api_key").(string))
	}
	if d.HasChange("models") {
		i.Models = expandAIRuntimeControlModels(d.Get("models").(*schema.Set))
	}

	// api_key is deliberately omitted from this log line; do not log the full
	// input struct, as it contains the provider's secret key.
	log.Printf("[DEBUG] UPDATE: AI Runtime Control provider connection input: id=%q base_url=%q models=%v", gofastly.ToValue(i.ID), gofastly.ToValue(i.BaseURL), i.Models)

	if _, err := providerconnection.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAIRuntimeControlProviderConnectionRead(ctx, d, meta)
}

func resourceFastlyAIRuntimeControlProviderConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := providerconnection.DeleteInput{
		ID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: AI Runtime Control provider connection input: %#v", i)

	if err := providerconnection.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func expandAIRuntimeControlModels(s *schema.Set) []string {
	models := make([]string, 0, s.Len())
	for _, m := range s.List() {
		models = append(models, m.(string))
	}
	return models
}
