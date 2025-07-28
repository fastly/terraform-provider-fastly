package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFThresholdsData_Config(t *testing.T) {
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

						// Check that we have at least one threshold
						var thresholdCount int
						for k := range a {
							if strings.HasPrefix(k, "thresholds.") && strings.HasSuffix(k, ".id") {
								thresholdCount++
							}
						}

						if thresholdCount == 0 {
							return fmt.Errorf("expected at least one threshold, got %d", thresholdCount)
						}

						// Check that threshold attributes are properly set
						for k, v := range a {
							if strings.HasSuffix(k, ".name") && v == "" {
								return fmt.Errorf("threshold name should not be empty")
							}
							if strings.HasSuffix(k, ".signal") && v == "" {
								return fmt.Errorf("threshold signal should not be empty")
							}
							if strings.HasSuffix(k, ".action") && v != "log" && v != "block" {
								return fmt.Errorf("threshold action should be 'log' or 'block', got: %s", v)
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFThresholdsConfig(h string) string {
	tf := `
resource "fastly_ngwaf_thresholds" "test_thesholds_workspace" {
  workspace_id = fastly_ngwaf_workspace.example.id
  name         = "test-threshold-%s"
  signal       = "SQLI"
  action       = "log"
  enabled      = true
  limit        = 10
  interval     = 3600
  duration     = 86400
  dont_notify  = false
}

data "fastly_ngwaf_thresholds" "sample" {
  workspace_id = fastly_ngwaf_workspace.example.id
}
`
	return fmt.Sprintf(tf, h, h, h)
}
