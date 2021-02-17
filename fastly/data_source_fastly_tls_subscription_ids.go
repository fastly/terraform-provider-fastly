package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSSubscriptionIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSSubscriptionIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Description: "IDs of available TLS subscriptions.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSSubscriptionIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	subscriptions, err := listTLSSubscriptions(conn)
	if err != nil {
		return err
	}

	var ids []string
	for _, subscription := range subscriptions {
		ids = append(ids, subscription.ID)
	}

	// 2.x upgrade note - `hashcode.String` was removed from the SDK
	// Code will need to be copied into this repository
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // if other filters are added to this data source, they should be included in this hashcode instead of the empty string
	if err := d.Set("ids", ids); err != nil {
		return err
	}
	return nil
}
