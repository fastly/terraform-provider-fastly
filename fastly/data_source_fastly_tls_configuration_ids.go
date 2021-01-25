package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSConfigurationIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSConfigurationIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "IDs of available TLS configurations",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSConfigurationIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	configurations, err := listTLSConfigurations(conn)
	if err != nil {
		return err
	}

	var ids []string
	for _, configuration := range configurations {
		ids = append(ids, configuration.ID)
	}

	// 2.x upgrade note - `hashcode.String` was removed from the SDK
	// Code will need to be copied into this repository
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // hashCode should include any filters set when they are added
	if err := d.Set("ids", ids); err != nil {
		return err
	}
	return nil
}
