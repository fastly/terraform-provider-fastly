package fastly

import (
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type dataSourceFastlyIPRangesResult struct {
	Addresses []string
}

func dataSourceFastlyIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyIPRangesRead,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv6_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyIPRangesRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	log.Printf("[DEBUG] Reading IP ranges")

	ipv4addresses, ipv6addresses, err := conn.AllIPs()

	if err != nil {
		return fmt.Errorf("Error listing IP ranges: %s", err)
	}

	d.SetId(hashcode.Strings(append(ipv4addresses, ipv6addresses...)))

	sort.Strings(ipv4addresses)
	sort.Strings(ipv6addresses)

	if err := d.Set("cidr_blocks", ipv4addresses); err != nil {
		return fmt.Errorf("Error setting ipv4 ranges: %s", err)
	}

	if err := d.Set("ipv6_cidr_blocks", ipv6addresses); err != nil {
		return fmt.Errorf("Error setting ipv6 ranges: %s", err)
	}

	return nil

}
