package fastly

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
)

func resourceFastlyAIRuntimeControlVirtualKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAIRuntimeControlVirtualKeyCreate,
		ReadContext:   resourceFastlyAIRuntimeControlVirtualKeyRead,
		UpdateContext: resourceFastlyAIRuntimeControlVirtualKeyUpdate,
		DeleteContext: resourceFastlyAIRuntimeControlVirtualKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (UTC) of when the virtual key was created.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the user who created the virtual key.",
			},
			"expires_at": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The expiration timestamp of the virtual key, in RFC 3339 format. Once set, this value cannot be unset, only changed.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
				DiffSuppressFunc: diffSuppressEquivalentRFC3339,
			},
			"last_used_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (UTC) of when the virtual key was last used.",
			},
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The AI model identifier, e.g. `claude-sonnet-4-20250514`.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable name for the virtual key.",
			},
			// The Terraform SDK reserves the field name "provider" on every
			// resource, so the API's `provider` field is exposed as
			// `provider_name`.
			"provider_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The AI model provider name, e.g. `Anthropic`.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the virtual key.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (UTC) of when the virtual key was last updated.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user creating the virtual key. Changing this forces a new virtual key to be created.",
			},
		},
	}
}

func resourceFastlyAIRuntimeControlVirtualKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := key.CreateInput{
		Name:     gofastly.ToPointer(d.Get("name").(string)),
		Model:    gofastly.ToPointer(d.Get("model").(string)),
		Provider: gofastly.ToPointer(d.Get("provider_name").(string)),
		UserID:   gofastly.ToPointer(d.Get("user_id").(string)),
	}

	if v, ok := d.GetOk("expires_at"); ok {
		expiresAt, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		i.ExpiresAt = &expiresAt
	}

	log.Printf("[DEBUG] CREATE: AI Runtime Control virtual key input: %#v", i)

	// The access token is returned only here and by the rotate endpoint, which
	// the provider does not expose. It is deliberately discarded rather than
	// persisted to state.
	vk, err := key.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vk.ID)

	return resourceFastlyAIRuntimeControlVirtualKeyRead(ctx, d, meta)
}

func resourceFastlyAIRuntimeControlVirtualKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := key.GetInput{
		KeyID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: AI Runtime Control virtual key input: %#v", i)

	vk, err := key.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] AI Runtime Control virtual key not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Virtual keys are soft deleted, so a deleted key may still be returned
	// rather than 404ing.
	if vk.DeletedAt != nil {
		log.Printf("[WARN] AI Runtime Control virtual key deleted '%s'", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", vk.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model", vk.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("provider_name", vk.Provider); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", vk.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", vk.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_by", vk.CreatedBy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", vk.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_id", vk.UserID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expires_at", vk.ExpiresAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_used_at", vk.LastUsedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyAIRuntimeControlVirtualKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// The API applies a partial update, so only send what actually changed.
	i := key.UpdateInput{
		KeyID: gofastly.ToPointer(d.Id()),
	}

	if d.HasChange("name") {
		i.Name = gofastly.ToPointer(d.Get("name").(string))
	}
	if d.HasChange("model") {
		i.Model = gofastly.ToPointer(d.Get("model").(string))
	}
	if d.HasChange("provider_name") {
		i.Provider = gofastly.ToPointer(d.Get("provider_name").(string))
	}
	if d.HasChange("expires_at") {
		if v, ok := d.GetOk("expires_at"); ok {
			expiresAt, err := time.Parse(time.RFC3339, v.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			i.ExpiresAt = &expiresAt
		}
	}

	log.Printf("[DEBUG] UPDATE: AI Runtime Control virtual key input: %#v", i)

	if _, err := key.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAIRuntimeControlVirtualKeyRead(ctx, d, meta)
}

func resourceFastlyAIRuntimeControlVirtualKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := key.DeleteInput{
		KeyID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: AI Runtime Control virtual key input: %#v", i)

	if err := key.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// diffSuppressEquivalentRFC3339 treats two RFC 3339 timestamps as equal when
// they denote the same instant, so that a UTC value returned by the API does
// not conflict with an equivalent offset-based value in the configuration.
func diffSuppressEquivalentRFC3339(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == "" || newValue == "" {
		return false
	}

	oldTime, err := time.Parse(time.RFC3339, oldValue)
	if err != nil {
		return false
	}
	newTime, err := time.Parse(time.RFC3339, newValue)
	if err != nil {
		return false
	}

	return oldTime.Equal(newTime)
}
