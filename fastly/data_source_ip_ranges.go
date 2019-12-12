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
		},
	}
}

func dataSourceFastlyIPRangesRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	log.Printf("[DEBUG] Reading IP ranges")

	addresses, err := conn.IPs()

	if err != nil {
		return fmt.Errorf("Error listing IP ranges: %s", err)
	}

	d.SetId(hashcode.Strings(addresses))

	sort.Strings(addresses)

	if err := d.Set("cidr_blocks", addresses); err != nil {
		return fmt.Errorf("Error setting ip ranges: %s", err)
	}

	return nil

}
