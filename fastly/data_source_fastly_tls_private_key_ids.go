package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceTLSPrivateKeyIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTLSPrivateKeyIdsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "List of IDs of the TLS private keys.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTLSPrivateKeyIdsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	keys, err := listTLSPrivateKeys(conn)
	if err != nil {
		return err
	}

	var ids []string
	for _, key := range keys {
		ids = append(ids, key.ID)
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
