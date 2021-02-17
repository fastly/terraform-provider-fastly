package fastly

import (
	"fmt"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSPlatformCertificateIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSPlatformCertificateIDsRead,
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

func dataSourceFastlyTLSPlatformCertificateIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	certificates, err := listPlatformTLSCertificates(conn)
	if err != nil {
		return err
	}

	var ids []string
	for _, certificate := range certificates {
		ids = append(ids, certificate.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // if other filters are added to this data source, they should be included in this hashcode instead of the empty string
	err = d.Set("ids", ids)
	if err != nil {
		return err
	}

	return nil
}
