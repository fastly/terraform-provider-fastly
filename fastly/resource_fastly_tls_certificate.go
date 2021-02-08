package fastly

import (
	"time"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTLSCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTLSCertificateCreate,
		Read:   resourceTLSCertificateRead,
		Update: resourceTLSCertificateUpdate,
		Delete: resourceTLSCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Human-readable name used to identify the certificate - defaults to the certificate's Common Name or first Subject Alternative Name entry",
				Optional:    true,
				Computed:    true,
			},
			"certificate_body": {
				Type:         schema.TypeString,
				Description:  "PEM-formatted certificate",
				Required:     true,
				ValidateFunc: validatePEMBlock("CERTIFICATE"),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was last updated",
				Computed:    true,
			},
			"issued_to": {
				Type:        schema.TypeString,
				Description: "The hostname for which a certificate was issued",
				Computed:    true,
			},
			"issuer": {
				Type:        schema.TypeString,
				Description: "The certificate authority that issued the certificate",
				Computed:    true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Description: "A recommendation from Fastly indicating the key associated with this certificate is in need of rotation",
				Computed:    true,
			},
			"serial_number": {
				Type:        schema.TypeString,
				Description: "A value assigned by the issuer that is unique to a certificate",
				Computed:    true,
			},
			"signature_algorithm": {
				Type:        schema.TypeString,
				Description: "The algorithm used to sign the certificate",
				Computed:    true,
			},
			"domains": {
				Type:        schema.TypeSet,
				Description: "All the domains (including wildcard domains) that are listed in any certificate's Subject Alternative Names (SAN) list",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTLSCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	input := &fastly.CreateCustomTLSCertificateInput{
		CertBlob: d.Get("certificate_body").(string),
	}

	if v, ok := d.GetOkExists("name"); ok {
		input.Name = v.(string)
	}

	output, err := conn.CreateCustomTLSCertificate(input)
	if err != nil {
		return err
	}

	d.SetId(output.ID)

	return resourceTLSCertificateRead(d, meta)
}

func resourceTLSCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	cert, err := conn.GetCustomTLSCertificate(&fastly.GetCustomTLSCertificateInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	var domains []string
	for _, domain := range cert.TLSDomains {
		domains = append(domains, domain.ID)
	}

	if err := d.Set("name", cert.Name); err != nil {
		return err
	}
	if err := d.Set("created_at", cert.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", cert.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("issued_to", cert.IssuedTo); err != nil {
		return err
	}
	if err := d.Set("issuer", cert.Issuer); err != nil {
		return err
	}
	if err := d.Set("replace", cert.Replace); err != nil {
		return err
	}
	if err := d.Set("serial_number", cert.SerialNumber); err != nil {
		return err
	}
	if err := d.Set("signature_algorithm", cert.SignatureAlgorithm); err != nil {
		return err
	}
	if err := d.Set("domains", domains); err != nil {
		return err
	}

	return nil
}

func resourceTLSCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	input := &fastly.UpdateCustomTLSCertificateInput{
		ID:       d.Id(),
		CertBlob: d.Get("certificate_body").(string),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = v.(string)
	}

	_, err := conn.UpdateCustomTLSCertificate(input)
	if err != nil {
		return err
	}

	return resourceTLSCertificateRead(d, meta)
}

func resourceTLSCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteCustomTLSCertificate(&fastly.DeleteCustomTLSCertificateInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	return nil
}
