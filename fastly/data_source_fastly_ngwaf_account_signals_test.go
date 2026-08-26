package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAccountSignals_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAccountSignalsConfig(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_account_signals.test"]
						a := r.Primary.Attributes

						// Check that we have the signals attribute
						signalCount := a["signals.#"]
						if signalCount == "" {
							return fmt.Errorf("expected signals attribute to be present")
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

func testAccFastlyDataSourceNGWAFAccountSignalsConfig() string {
	return `
data "fastly_ngwaf_account_signals" "test" {
}
`
}
