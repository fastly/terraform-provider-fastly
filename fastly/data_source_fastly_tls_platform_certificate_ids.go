package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSPlatformCertificateIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSPlatformCertificateIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "IDs of Platform TLS certificates",
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

	// 2.x upgrade note - `hashcode.String` was removed from the SDK
	// Code will need to be copied into this repository
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // hashCode should include any filters set when they are added
	err = d.Set("ids", ids)
	if err != nil {
		return err
	}

	return nil
}
