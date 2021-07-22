package fastly

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyIPRanges(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyIPRangesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyIPRanges("data.fastly_ip_ranges.some"),
				),
			},
		},
	})
}

func testAccFastlyIPRanges(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		var (
			cidrBlockSize     int
			ipv6cidrBlockSize int
			err               error
		)

		if cidrBlockSize, err = strconv.Atoi(a["cidr_blocks.#"]); err != nil {
			return err
		}

		if ipv6cidrBlockSize, err = strconv.Atoi(a["ipv6_cidr_blocks.#"]); err != nil {
			return err
		}

		if cidrBlockSize < 10 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		if 0 >= ipv6cidrBlockSize {
			return fmt.Errorf("ipv6_cidr_blocks are missing")
		}

		var cidrBlocks sort.StringSlice = make([]string, cidrBlockSize)

		for i := range make([]string, cidrBlockSize) {

			block := a[fmt.Sprintf("cidr_blocks.%d", i)]

			if _, _, err := net.ParseCIDR(block); err != nil {
				return fmt.Errorf("malformed CIDR block %s: %s", block, err)
			}

			cidrBlocks[i] = block

		}

		if !sort.IsSorted(cidrBlocks) {
			return fmt.Errorf("unexpected order of cidr_blocks: %s", cidrBlocks)
		}

		var ipv6cidrBlocks sort.StringSlice = make([]string, ipv6cidrBlockSize)

		for j := range make([]string, ipv6cidrBlockSize) {

			block := a[fmt.Sprintf("ipv6_cidr_blocks.%d", j)]

			if _, _, err := net.ParseCIDR(block); err != nil {
				return fmt.Errorf("malformed CIDR block %s: %s", block, err)
			}

			ipv6cidrBlocks[j] = block
		}

		if !sort.IsSorted(ipv6cidrBlocks) {
			return fmt.Errorf("unexpected order of ipv6_cidr_blocks: %s", ipv6cidrBlocks)
		}

		return nil
	}
}

const testAccFastlyIPRangesConfig = `
provider "fastly" {
  alias   = "noauth"
  api_key = ""
  no_auth = true
}
data "fastly_ip_ranges" "some" {
  provider = fastly.noauth
}
`
