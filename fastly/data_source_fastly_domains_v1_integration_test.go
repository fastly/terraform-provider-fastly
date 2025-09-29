package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceDomainsV1_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceDomainsV1Config(),
				Check: resource.ComposeTestCheckFunc(
					// Basic structure validation
					resource.TestCheckResourceAttrSet("data.fastly_domains_v1.example", "domains.#"),
					resource.TestCheckResourceAttrSet("data.fastly_domains_v1.example", "total"),

					// Advanced validation function to check data quality
					testAccDomainsV1DataSourceState("data.fastly_domains_v1.example"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceDomainsV1Config() string {
	return `
data "fastly_domains_v1" "example" {
}
`
}

func testAccDomainsV1DataSourceState(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[resourceName]
		a := r.Primary.Attributes

		domainsCount, err := strconv.Atoi(a["domains.#"])
		if err != nil {
			return fmt.Errorf("domains.# is not a valid number: %v", a["domains.#"])
		}

		totalCount, err := strconv.Atoi(a["total"])
		if err != nil {
			return fmt.Errorf("total is not a valid number: %v", a["total"])
		}

		// Verify total matches domains count
		if domainsCount != totalCount {
			return fmt.Errorf("domains count (%d) doesn't match total (%d)", domainsCount, totalCount)
		}

		// If domains exist, verify structure of first domain
		if domainsCount > 0 {
			if fqdn, exists := a["domains.0.fqdn"]; !exists || fqdn == "" {
				return fmt.Errorf("first domain missing or has empty fqdn")
			}
			if id, exists := a["domains.0.id"]; !exists || id == "" {
				return fmt.Errorf("first domain missing or has empty id")
			}
			// service_id isn't always returned, so we don't validate it
		}

		return nil
	}
}
