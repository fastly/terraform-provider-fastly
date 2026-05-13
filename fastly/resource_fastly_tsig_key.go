package fastly

import (
	"context"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/tsigkeys"
)

func resourceFastlyTSIGKeys() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTSIGKeysCreate,
		ReadContext:   resourceFastlyTSIGKeysRead,
		UpdateContext: resourceFastlyTSIGKeysUpdate,
		DeleteContext: resourceFastlyTSIGKeysDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"algorithm": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The algorithm of the TSIG key. One of: `hmac-sha224`, `hmac-sha256`, `hmac-sha384`,  `hmac-sha512`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"hmac-sha224", "hmac-sha256", "hmac-sha384", "hmac-sha512"},
					false,
				)),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for your TSIG Key.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the TSIG Key.",
				// These constraints  may change in the future - so we are using simple validation for now.
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringMatch(regexp.MustCompile(`^\S+$`), "must not contain spaces"),
				),
			},
			"secret": {
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				Description:      "The Base64 encoded secret key.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsBase64),
			},
		},
	}
}

func resourceFastlyTSIGKeysCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var input tsigkeys.CreateInput
	if v, ok := d.GetOk("algorithm"); ok {
		input.Algorithm = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		input.Name = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("secret"); ok {
		input.Secret = gofastly.ToPointer(v.(string))
	}

	data, err := tsigkeys.Create(ctx, conn, &input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*data.ID)
	return resourceFastlyTSIGKeysRead(ctx, d, meta)
}

func resourceFastlyTSIGKeysRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TSIG Key for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	input := &tsigkeys.GetInput{
		TSIGKeyID: gofastly.ToPointer(d.Id()),
	}

	data, err := tsigkeys.Get(ctx, conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("algorithm", data.Algorithm); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", data.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", data.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyTSIGKeysUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &tsigkeys.UpdateInput{
		TSIGKeyID: gofastly.ToPointer(d.Id()),
	}

	if v, ok := d.GetOk("algorithm"); ok {
		input.Algorithm = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.NewNullable(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		input.Name = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("secret"); ok {
		input.Secret = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Updating TSIG Key: %#v", input)
	_, err := tsigkeys.Update(ctx, conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyTSIGKeysRead(ctx, d, meta)
}

func resourceFastlyTSIGKeysDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	input := &tsigkeys.DeleteInput{
		TSIGKeyID: gofastly.ToPointer(d.Id()),
	}
	log.Printf("[DEBUG] Deleting TSIG Key: %#v", input)
	err := tsigkeys.Delete(ctx, conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}
