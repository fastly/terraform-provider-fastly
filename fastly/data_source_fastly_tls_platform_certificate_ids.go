package fastly

import (
	"context"
	"fmt"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSPlatformCertificateIDs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTLSPlatformCertificateIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Description: "List of IDs corresponding to Platform TLS certificates.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSPlatformCertificateIDsRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	certificates, err := listPlatformTLSCertificates(conn)
	if err != nil {
		return diag.FromErr(err)
	}

	var ids []string
	for _, certificate := range certificates {
		ids = append(ids, certificate.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // if other filters are added to this data source, they should be included in this hashcode instead of the empty string
	err = d.Set("ids", ids)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
