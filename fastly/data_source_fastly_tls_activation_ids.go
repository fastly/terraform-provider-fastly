package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSActivationIds() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSActivationIDsRead,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of TLS certificate used to filter activations",
			},
			"ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "List of IDs of the TLS Activations.",
			},
		},
	}
}

func dataSourceFastlyTLSActivationIDsRead(d *schema.ResourceData, meta interface{}) error {
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

	var ids []string
	for _, activation := range activations {
		ids = append(ids, activation.ID)
	}

	// 2.x upgrade note - `hashcode.String` was removed from the SDK
	// Code will need to be copied into this repository
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
	d.SetId(fmt.Sprintf("%d", hashcode.String(certificateID)))
	err := d.Set("ids", ids)
	if err != nil {
		return err
	}

	return nil
}
