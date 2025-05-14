package fastly

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v10/fastly"
)

func resourceFastlyTLSCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSCertificateCreate,
		ReadContext:   resourceFastlyTLSCertificateRead,
		UpdateContext: resourceFastlyTLSCertificateUpdate,
		DeleteContext: resourceFastlyTLSCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:             schema.TypeString,
				Description:      "PEM-formatted certificate, optionally including any intermediary certificates.",
				Required:         true,
				ValidateDiagFunc: validatePEMBlocks("CERTIFICATE"),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was created.",
				Computed:    true,
			},
			"domains": {
				Type:        schema.TypeSet,
				Description: "All the domains (including wildcard domains) that are listed in the certificate's Subject Alternative Names (SAN) list.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"issued_to": {
				Type:        schema.TypeString,
				Description: "The hostname for which a certificate was issued.",
				Computed:    true,
			},
			"issuer": {
				Type:        schema.TypeString,
				Description: "The certificate authority that issued the certificate.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Human-readable name used to identify the certificate. Defaults to the certificate's Common Name or first Subject Alternative Name entry.",
				Optional:    true,
				Computed:    true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Description: "A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.",
				Computed:    true,
			},
			"serial_number": {
				Type:        schema.TypeString,
				Description: "A value assigned by the issuer that is unique to a certificate.",
				Computed:    true,
			},
			"signature_algorithm": {
				Type:        schema.TypeString,
				Description: "The algorithm used to sign the certificate.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was last updated.",
				Computed:    true,
			},
		},
	}
}

func resourceFastlyTLSCertificateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.CreateCustomTLSCertificateInput{
		CertBlob: d.Get("certificate_body").(string),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = v.(string)
	}

	output, err := conn.CreateCustomTLSCertificate(input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.ID)

	return resourceFastlyTLSCertificateRead(ctx, d, meta)
}

func resourceFastlyTLSCertificateRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TLS Certificate Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	var diags diag.Diagnostics

	cert, err := conn.GetCustomTLSCertificate(&fastly.GetCustomTLSCertificateInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var domains []string
	for _, domain := range cert.Domains {
		domains = append(domains, domain.ID)
	}

	if cert.Replace {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fastly recommends that this certificate (%s) be replaced", cert.ID),
		})
	}

	if err := d.Set("name", cert.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", cert.CreatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", cert.UpdatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("issued_to", cert.IssuedTo); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("issuer", cert.Issuer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("replace", cert.Replace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("serial_number", cert.SerialNumber); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("signature_algorithm", cert.SignatureAlgorithm); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domains", domains); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFastlyTLSCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.UpdateCustomTLSCertificateInput{
		ID:       d.Id(),
		CertBlob: d.Get("certificate_body").(string),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = v.(string)
	}

	_, err := conn.UpdateCustomTLSCertificate(input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyTLSCertificateRead(ctx, d, meta)
}

func resourceFastlyTLSCertificateDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteCustomTLSCertificate(&fastly.DeleteCustomTLSCertificateInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
