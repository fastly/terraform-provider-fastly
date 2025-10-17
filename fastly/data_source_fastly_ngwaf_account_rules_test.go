package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAccountRules_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAccountRulesConfig(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_account_rules.test"]
						a := r.Primary.Attributes

						// Check that we have the rules attribute
						ruleCount := a["rules.#"]
						if ruleCount == "" {
							return fmt.Errorf("expected rules attribute to be present")
						}

						// If there are rules, check that the first one has an ID
						if ruleCount != "0" {
							ruleID := a["rules.0.id"]
							if ruleID == "" {
								return fmt.Errorf("expected rule to have an ID, got empty string")
							}

							// Check that rule has required fields
							ruleType := a["rules.0.type"]
							if ruleType == "" {
								return fmt.Errorf("expected rule to have a type, got empty string")
							}

							// Check enabled is present
							if _, ok := a["rules.0.enabled"]; !ok {
								return fmt.Errorf("expected rule to have enabled field")
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAccountRulesConfig() string {
	return `
data "fastly_ngwaf_account_rules" "test" {
}
`
}
