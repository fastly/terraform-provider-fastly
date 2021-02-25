package fastly

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"sort"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type dataSourceFastlyIPRangesResult struct {
	Addresses []string
}

func dataSourceFastlyIPRanges() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyIPRangesRead,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The lexically ordered list of ipv4 CIDR blocks.",
			},
			"ipv6_cidr_blocks": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The lexically ordered list of ipv6 CIDR blocks.",
			},
		},
	}
}

func dataSourceFastlyIPRangesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*FastlyClient).conn

	log.Printf("[DEBUG] Reading IP ranges")

	ipv4addresses, ipv6addresses, err := conn.AllIPs()

	if err != nil {
		return diag.Errorf("Error listing IP ranges: %s", err)
	}

	d.SetId(hashcode.Strings(append(ipv4addresses, ipv6addresses...)))

	sort.Strings(ipv4addresses)
	sort.Strings(ipv6addresses)

	if err := d.Set("cidr_blocks", ipv4addresses); err != nil {
		return diag.Errorf("Error setting ipv4 ranges: %s", err)
	}

	if err := d.Set("ipv6_cidr_blocks", ipv6addresses); err != nil {
		return diag.Errorf("Error setting ipv6 ranges: %s", err)
	}

	return nil

}
