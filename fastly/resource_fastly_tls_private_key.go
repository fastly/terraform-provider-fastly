package fastly

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

func resourceFastlyTLSPrivateKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSPrivateKeyCreate,
		ReadContext:   resourceFastlyTLSPrivateKeyRead,
		DeleteContext: resourceFastlyTLSPrivateKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time-stamp (GMT) when the private key was created.",
			},
			"key_length": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The key length used to generate the private key.",
			},
			"key_pem": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Private key in PEM format.",
				Sensitive:   !DisplaySensitiveFields,
			},
			"key_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The algorithm used to generate the private key. Must be RSA.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Customisable name of the private key.",
			},
			"public_key_sha1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Useful for safely identifying the key.",
			},
			"replace": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Fastly recommends replacing this private key.",
			},
		},
	}
}

func resourceFastlyTLSPrivateKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	privateKey, err := conn.CreatePrivateKey(ctx, &gofastly.CreatePrivateKeyInput{
		Key:  d.Get("key_pem").(string),
		Name: d.Get("name").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(privateKey.ID)

	return resourceFastlyTLSPrivateKeyRead(ctx, d, meta)
}

func resourceFastlyTLSPrivateKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TLS Private Key Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	var diags diag.Diagnostics

	privateKey, err := conn.GetPrivateKey(ctx, &gofastly.GetPrivateKeyInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if privateKey.Replace {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fastly recommends that this private key (%s) be replaced", privateKey.ID),
		})
	}

	err = d.Set("name", privateKey.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("created_at", privateKey.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("key_length", privateKey.KeyLength)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("key_type", privateKey.KeyType)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("replace", privateKey.Replace)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("public_key_sha1", privateKey.PublicKeySHA1)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFastlyTLSPrivateKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeletePrivateKey(ctx, &gofastly.DeletePrivateKeyInput{
		ID: d.Id(),
	})

	return diag.FromErr(err)
}
