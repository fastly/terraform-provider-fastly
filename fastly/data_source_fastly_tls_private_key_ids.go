package fastly

import (
	"fmt"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSPrivateKeyIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSPrivateKeyIDsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Description: "List of IDs of the TLS private keys.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSPrivateKeyIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	keys, err := listTLSPrivateKeys(conn)
	if err != nil {
		return err
	}

	var ids []string
	for _, key := range keys {
		ids = append(ids, key.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // if other filters are added to this data source, they should be included in this hashcode instead of the empty string
	if err := d.Set("ids", ids); err != nil {
		return err
	}

	return nil
}
