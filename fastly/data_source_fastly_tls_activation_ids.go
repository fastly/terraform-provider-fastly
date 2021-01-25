package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSActivationIds() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSActivationIdsRead,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of TLS certificate used to filter activations",
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceFastlyTLSActivationIdsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var certificateID string

	if v, ok := d.GetOk("certificate_id"); ok {
		certificateID = v.(string)
	}

	var activations []*fastly.TLSActivation
	pageNumber := 1
	for {
		list, err := conn.ListTLSActivations(&fastly.ListTLSActivationsInput{
			FilterTLSCertificateID: certificateID,
			PageNumber:             pageNumber,
			PageSize:               10,
		})
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		activations = append(activations, list...)
	}

	if len(activations) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
	}

	activationIds := make([]string, 0)

	for _, activation := range activations {
		activationIds = append(activationIds, activation.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(certificateID)))
	d.Set("ids", activationIds)

	return nil
}
