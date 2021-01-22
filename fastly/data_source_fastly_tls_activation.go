package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func dataSourceFastlyTLSActivation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSActivationRead,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of TLS certificate to enable.",
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of TLS configuration to be used to terminate TLS traffic, or use the default one if missing.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Domain to enable TLS traffic on.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Time-stamp (GMT) when TLS was enabled.",
			},
		},
	}
}

func dataSourceFastlyTLSActivationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var filters fastly.ListTLSActivationsInput

	if v, ok := d.GetOk("certificate_id"); ok {
		filters.FilterTLSCertificateID = v.(string)
	}
	if v, ok := d.GetOk("configuration_id"); ok {
		filters.FilterTLSConfigurationID = v.(string)
	}
	if v, ok := d.GetOk("domain"); ok {
		filters.FilterTLSDomainID = v.(string)
	}

	activations, err := conn.ListTLSActivations(&filters)
	if err != nil {
		return err
	}

	if len(activations) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
	}

	if len(activations) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please change to a more specific search criteria")
	}

	activation := activations[0]

	d.SetId(activation.ID)
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

	return err
}
