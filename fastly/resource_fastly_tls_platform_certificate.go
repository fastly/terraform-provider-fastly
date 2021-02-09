package fastly

import (
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTLSPlatformCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTLSPlatformCertificateCreate,
		Read:   resourceTLSPlatformCertificateRead,
		Update: resourceTLSPlatformCertificateUpdate,
		Delete: resourceTLSPlatformCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:         schema.TypeString,
				Description:  "PEM-formatted certificate.",
				Required:     true,
				ValidateFunc: validatePEMBlock("CERTIFICATE"),
			},
			"intermediates_blob": {
				Type:         schema.TypeString,
				Description:  "PEM-formatted certificate chain from the `certificate_body` to its root.",
				Required:     true,
				ValidateFunc: validatePEMBlocks("CERTIFICATE"),
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

func resourceTLSPlatformCertificateCreate(d *schema.ResourceData, meta interface{}) error {
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
		return err
	}

	d.SetId(certificate.ID)

	return resourceTLSPlatformCertificateRead(d, meta)
}

func resourceTLSPlatformCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	certificate, err := conn.GetBulkCertificate(&fastly.GetBulkCertificateInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	var domains []string
	for _, domain := range certificate.Domains {
		domains = append(domains, domain.ID)
	}

	if err := d.Set("configuration_id", certificate.Configurations[0].ID); err != nil {
		return err
	}
	if err := d.Set("not_after", certificate.NotAfter.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("not_before", certificate.NotBefore.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("created_at", certificate.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", certificate.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("replace", certificate.Replace); err != nil {
		return err
	}
	if err := d.Set("domains", domains); err != nil {
		return err
	}

	return nil
}

func resourceTLSPlatformCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	_, err := conn.UpdateBulkCertificate(&fastly.UpdateBulkCertificateInput{
		ID:                d.Id(),
		CertBlob:          d.Get("certificate_body").(string),
		IntermediatesBlob: d.Get("intermediates_blob").(string),
		AllowUntrusted:    d.Get("allow_untrusted_root").(bool),
	})
	if err != nil {
		return err
	}

	return resourceTLSPlatformCertificateRead(d, meta)
}

func resourceTLSPlatformCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	if err := conn.DeleteBulkCertificate(&fastly.DeleteBulkCertificateInput{
		ID: d.Id(),
	}); err != nil {
		return err
	}

	return nil
}
