package fastly

import (
	"context"
	"fmt"

	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSActivationIds() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTLSActivationIDsRead,
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
				Description: "List of IDs of the TLS Activations.",
			},
		},
	}
}

func dataSourceFastlyTLSActivationIDsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

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
			return diag.FromErr(err)
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

	d.SetId(fmt.Sprintf("%d", hashcode.String(certificateID)))
	err := d.Set("ids", ids)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
