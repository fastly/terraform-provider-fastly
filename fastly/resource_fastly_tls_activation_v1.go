package fastly

import (
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTLSActivationV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceTLSActivationV1Create,
		Read:   resourceTLSActivationV1Read,
		Update: resourceTLSActivationV1Update,
		Delete: resourceTLSActivationV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTLSActivationV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	ta, err := conn.CreateTLSActivation(&gofastly.CreateTLSActivationInput{
		Certificate:   &gofastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
		Configuration: &gofastly.TLSConfiguration{ID: d.Get("configuration_id").(string)},
		Domain:        &gofastly.TLSDomain{ID: d.Get("domain_id").(string)},
	})

	if err != nil {
		return err
	}

	d.SetId(ta.ID)

	return resourceTLSActivationV1Read(d, meta)
}

func resourceTLSActivationV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	ta, err := conn.GetTLSActivation(&gofastly.GetTLSActivationInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("certificate_id", ta.Certificate.ID)
	d.Set("configuration_id", ta.Configuration.ID)
	d.Set("domain_id", ta.Domain.ID)
	d.Set("created_at", ta.CreatedAt.String())

	return nil
}

func resourceTLSActivationV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Update Name and/or Role.
	if d.HasChange("certificate_id") {
		_, err := conn.UpdateTLSActivation(&gofastly.UpdateTLSActivationInput{
			ID:          d.Id(),
			Certificate: &gofastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
		})

		if err != nil {
			return err
		}
	}

	return resourceTLSActivationV1Read(d, meta)
}

func resourceTLSActivationV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteTLSActivation(&gofastly.DeleteTLSActivationInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	return nil
}
