package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFWorkspaceRules_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFWorkspaceRulesConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_workspace_rules.test"]
						a := r.Primary.Attributes

						// Check that we have the rules attribute
						ruleCount := a["rules.#"]
						if ruleCount == "" {
							return fmt.Errorf("expected rules attribute to be present")
						}

						// Check that workspace_id is set
						workspaceID := a["workspace_id"]
						if workspaceID == "" {
							return fmt.Errorf("expected workspace_id to be set")
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

							// Check that rule has description
							ruleDesc := a["rules.0.description"]
							if ruleDesc == "" {
								return fmt.Errorf("expected rule to have a description, got empty string")
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFWorkspaceRulesConfig(h string) string {
	return fmt.Sprintf(`
%s

data "fastly_ngwaf_workspace_rules" "test" {
  workspace_id = fastly_ngwaf_workspace.example.id
}
`, testAccNGWAFWorkspaceConfig(fmt.Sprintf("tf_%s", h)))
}
