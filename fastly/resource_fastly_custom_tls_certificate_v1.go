package fastly

import (
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCustomTLSCertificateV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomTLSCertificateV1Create,
		Read:   resourceCustomTLSCertificateV1Read,
		Update: resourceCustomTLSCertificateV1Update,
		Delete: resourceCustomTLSCertificateV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cert_blob": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed
			"issued_to": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signature_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domains": { // TODO Should this be a nested object?
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceCustomTLSCertificateV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	ta, err := conn.CreateCustomTLSCertificate(&gofastly.CreateCustomTLSCertificateInput{
		CertBlob: d.Get("cert_blob").(string),
		Name:     d.Get("name").(string),
	})

	if err != nil {
		return err
	}

	d.SetId(ta.ID)

	return resourceCustomTLSCertificateV1Read(d, meta)
}

func resourceCustomTLSCertificateV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	cc, err := conn.GetCustomTLSCertificate(&gofastly.GetCustomTLSCertificateInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	var domains []string
	for _, d := range cc.TLSDomains {
		domains = append(domains, d.ID)
	}

	d.Set("name", cc.Name)
	d.Set("issued_to", cc.IssuedTo)
	d.Set("issuer", cc.Issuer)
	d.Set("not_after", cc.NotAfter.String())
	d.Set("not_before", cc.NotBefore.String())
	d.Set("replace", cc.Replace)
	d.Set("serial_number", cc.SerialNumber)
	d.Set("signature_algorithm", cc.SignatureAlgorithm)
	d.Set("created_at", cc.CreatedAt.String())
	d.Set("updated_at", cc.UpdatedAt.String())
	d.Set("domains", domains)

	return nil
}

func resourceCustomTLSCertificateV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Update Name and/or CertBlob.
	if d.HasChange("cert_blob") || d.HasChange("name") { // TODO is this the best way to do this?
		_, err := conn.UpdateCustomTLSCertificate(&gofastly.UpdateCustomTLSCertificateInput{
			ID:       d.Id(),
			CertBlob: d.Get("cert_blob").(string),
			Name:     d.Get("name").(string),
		})

		if err != nil {
			return err
		}
	}

	return resourceCustomTLSCertificateV1Read(d, meta)
}

func resourceCustomTLSCertificateV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteCustomTLSCertificate(&gofastly.DeleteCustomTLSCertificateInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	return nil
}
