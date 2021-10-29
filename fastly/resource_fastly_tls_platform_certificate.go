package fastly

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"time"

	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyTLSPlatformCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSPlatformCertificateCreate,
		ReadContext:   resourceFastlyTLSPlatformCertificateRead,
		UpdateContext: resourceFastlyTLSPlatformCertificateUpdate,
		DeleteContext: resourceFastlyTLSPlatformCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:             schema.TypeString,
				Description:      "PEM-formatted certificate.",
				Required:         true,
				ValidateDiagFunc: validatePEMBlock("CERTIFICATE"),
			},
			"intermediates_blob": {
				Type:             schema.TypeString,
				Description:      "PEM-formatted certificate chain from the `certificate_body` to its root.",
				Required:         true,
				ValidateDiagFunc: validatePEMBlocks("CERTIFICATE"),
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Description: "ID of TLS configuration to be used to terminate TLS traffic.",
				Required:    true,
				ForceNew:    true,
			},
			"allow_untrusted_root": {
				Type:        schema.TypeBool,
				Description: "Disable checking whether the root of the certificate chain is trusted. Useful for development purposes to allow use of self-signed CAs. Defaults to false. Write-only on create.",
				Optional:    true,
				Default:     false,
			},
			"not_after": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate will expire.",
				Computed:    true,
			},
			"not_before": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate will become valid.",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was created.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was last updated.",
				Computed:    true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Description: "A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.",
				Computed:    true,
			},
			"domains": {
				Type:        schema.TypeSet,
				Description: "All the domains (including wildcard domains) that are listed in any certificate's Subject Alternative Names (SAN) list.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceFastlyTLSPlatformCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	input := &fastly.CreateBulkCertificateInput{
		CertBlob:          d.Get("certificate_body").(string),
		IntermediatesBlob: d.Get("intermediates_blob").(string),
		Configurations: []*fastly.TLSConfiguration{{
			ID: d.Get("configuration_id").(string),
		}},
		AllowUntrusted: d.Get("allow_untrusted_root").(bool),
	}

	certificate, err := conn.CreateBulkCertificate(input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(certificate.ID)

	return resourceFastlyTLSPlatformCertificateRead(ctx, d, meta)
}

func resourceFastlyTLSPlatformCertificateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	var diags diag.Diagnostics

	certificate, err := conn.GetBulkCertificate(&fastly.GetBulkCertificateInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var domains []string
	for _, domain := range certificate.Domains {
		domains = append(domains, domain.ID)
	}

	if certificate.Replace {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fastly recommends that this certificate (%s) be replaced", certificate.ID),
		})
	}

	if err := d.Set("configuration_id", certificate.Configurations[0].ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("not_after", certificate.NotAfter.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("not_before", certificate.NotBefore.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", certificate.CreatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", certificate.UpdatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("replace", certificate.Replace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domains", domains); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFastlyTLSPlatformCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	_, err := conn.UpdateBulkCertificate(&fastly.UpdateBulkCertificateInput{
		ID:                d.Id(),
		CertBlob:          d.Get("certificate_body").(string),
		IntermediatesBlob: d.Get("intermediates_blob").(string),
		AllowUntrusted:    d.Get("allow_untrusted_root").(bool),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyTLSPlatformCertificateRead(ctx, d, meta)
}

func resourceFastlyTLSPlatformCertificateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	if err := conn.DeleteBulkCertificate(&fastly.DeleteBulkCertificateInput{
		ID: d.Id(),
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
