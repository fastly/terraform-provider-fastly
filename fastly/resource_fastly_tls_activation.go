package fastly

import (
	"log"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyTLSActivation() *schema.Resource {
	return &schema.Resource{
		Create: resourceFastlyTLSActivationCreate,
		Read:   resourceFastlyTLSActivationRead,
		Update: resourceFastlyTLSActivationUpdate,
		Delete: resourceFastlyTLSActivationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of certificate to use. Must have the `domain` specified in the certificate's Subject Alternative Names.",
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "ID of TLS configuration to be used to terminate TLS traffic, or use the default one if missing.",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Domain to enable TLS on. Must be assigned to an existing Fastly Service.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time-stamp (GMT) when TLS was enabled.",
			},
		},
	}
}

func resourceFastlyTLSActivationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var configuration *fastly.TLSConfiguration
	if v, ok := d.GetOk("configuration_id"); ok {
		configuration = &fastly.TLSConfiguration{ID: v.(string)}
	}

	activation, err := conn.CreateTLSActivation(&fastly.CreateTLSActivationInput{
		Certificate:   &fastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
		Configuration: configuration,
		Domain:        &fastly.TLSDomain{ID: d.Get("domain").(string)},
	})
	if err != nil {
		return err
	}

	d.SetId(activation.ID)

	return resourceFastlyTLSActivationRead(d, meta)
}

func resourceFastlyTLSActivationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	activation, err := conn.GetTLSActivation(&fastly.GetTLSActivationInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	err = d.Set("certificate_id", activation.Certificate.ID)
	if err != nil {
		return err
	}
	err = d.Set("configuration_id", activation.Configuration.ID)
	if err != nil {
		return err
	}
	err = d.Set("domain", activation.Domain.ID)
	if err != nil {
		return err
	}
	err = d.Set("created_at", activation.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}

func resourceFastlyTLSActivationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	_, err := conn.UpdateTLSActivation(&fastly.UpdateTLSActivationInput{
		ID:          d.Id(),
		Certificate: &fastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
	})
	if err != nil {
		return err
	}

	return resourceFastlyTLSActivationRead(d, meta)
}

func resourceFastlyTLSActivationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteTLSActivation(&fastly.DeleteTLSActivationInput{
		ID: d.Id(),
	})
	if err != nil {
		if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.IsNotFound() {
			log.Printf("[WARN] Error deleting TLS activation (%s), not found. Was a TLS subscription enabled on the same domain?\n", d.Id())
			return nil
		}
		return err
	}

	return nil
}
