package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFThresholds_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFThresholdsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_thresholds.sample"]
						a := r.Primary.Attributes

						// Check that we have at least one threshold with an ID
						thresholdCount := a["thresholds.#"]
						if thresholdCount == "0" {
							return fmt.Errorf("expected at least one threshold, got %s", thresholdCount)
						}

						// Check that the first threshold has an ID
						thresholdID := a["thresholds.0.id"]
						if thresholdID == "" {
							return fmt.Errorf("expected threshold to have an ID, got empty string")
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFThresholdsConfig(h string) string {
	return fmt.Sprintf(`
%s

resource "fastly_ngwaf_thresholds" "test_threshold" {
  workspace_id = fastly_ngwaf_workspace.example.id
  name         = "Threshold TF Test"
  signal       = "SQLI"
  action       = "log"
  enabled      = true
  limit        = 10
  interval     = 3600
  duration     = 86400
  dont_notify  = false
}

data "fastly_ngwaf_thresholds" "sample" {
  depends_on = [
    fastly_ngwaf_thresholds.test_threshold
  ]
  workspace_id = fastly_ngwaf_workspace.example.id
}
`, testAccNGWAFWorkspaceConfig(fmt.Sprintf("tf_%s", h)))
}
