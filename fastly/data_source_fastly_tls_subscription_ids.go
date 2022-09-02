package fastly

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSSubscriptionIDs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTLSSubscriptionIDsRead,
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

func dataSourceFastlyTLSSubscriptionIDsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	subscriptions, err := listTLSSubscriptions(conn)
	if err != nil {
		return diag.FromErr(err)
	}

	var ids []string
	for _, subscription := range subscriptions {
		ids = append(ids, subscription.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(""))) // if other filters are added to this data source, they should be included in this hashcode instead of the empty string
	if err := d.Set("ids", ids); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
