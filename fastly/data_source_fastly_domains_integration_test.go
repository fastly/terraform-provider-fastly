package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceDomains_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceDomainsConfig(),
				Check: resource.ComposeTestCheckFunc(
					// Basic structure validation
					resource.TestCheckResourceAttrSet("data.fastly_domains.example", "domains.#"),
					resource.TestCheckResourceAttrSet("data.fastly_domains.example", "total"),

					// Advanced validation function to check data quality
					testAccDomainsDataSourceState("data.fastly_domains.example"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceDomainsConfig() string {
	return `
data "fastly_domains" "example" {
}
`
}

func testAccDomainsDataSourceState(resourceName string) resource.TestCheckFunc {
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
