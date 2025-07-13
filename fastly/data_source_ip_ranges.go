package fastly

import (
	"context"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

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

func dataSourceFastlyIPRangesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading IP ranges")

	ipv4addresses, ipv6addresses, err := conn.AllIPs(ctx)
	if err != nil {
		return diag.Errorf("error listing IP ranges: %s", err)
	}

	s, err := hashcode.Strings(append(ipv4addresses, ipv6addresses...))
	if err != nil {
		return diag.Errorf("error hashing IP ranges for internal state management: %s", err)
	}

	d.SetId(s)

	sort.Strings(ipv4addresses)
	sort.Strings(ipv6addresses)

	if err := d.Set("cidr_blocks", ipv4addresses); err != nil {
		return diag.Errorf("error setting ipv4 ranges: %s", err)
	}

	if err := d.Set("ipv6_cidr_blocks", ipv6addresses); err != nil {
		return diag.Errorf("error setting ipv6 ranges: %s", err)
	}

	return nil
}
