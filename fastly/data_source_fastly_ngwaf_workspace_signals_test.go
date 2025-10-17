package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFWorkspaceSignals_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFWorkspaceSignalsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_workspace_signals.test"]
						a := r.Primary.Attributes

						// Check that we have the signals attribute
						signalCount := a["signals.#"]
						if signalCount == "" {
							return fmt.Errorf("expected signals attribute to be present")
						}

						// Check that workspace_id is set
						workspaceID := a["workspace_id"]
						if workspaceID == "" {
							return fmt.Errorf("expected workspace_id to be set")
						}

						// If there are signals, check that the first one has an ID
						if signalCount != "0" {
							signalID := a["signals.0.id"]
							if signalID == "" {
								return fmt.Errorf("expected signal to have an ID, got empty string")
							}

							// Check that signal has required fields
							signalName := a["signals.0.name"]
							if signalName == "" {
								return fmt.Errorf("expected signal to have a name, got empty string")
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFWorkspaceSignalsConfig(h string) string {
	return fmt.Sprintf(`
%s

data "fastly_ngwaf_workspace_signals" "test" {
  workspace_id = fastly_ngwaf_workspace.example.id
}
`, testAccNGWAFWorkspaceConfig(fmt.Sprintf("tf_%s", h)))
}
