package fastly

import (
	"fmt"
	"slices"
	"strings"
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

						// Debug: print all attributes
						fmt.Printf("All attributes: %+v\n", a)

						expectedName := "Threshold TF Test"
						var (
							found      int
							gotNames   []string
							gotSignals []string
							gotActions []string
						)

						// Check for threshold attributes like:
						// "thresholds.0.name": "Threshold TF Test"
						// "thresholds.0.signal": "SQLI"
						// "thresholds.0.action": "log"
						for k, v := range a {
							if strings.HasSuffix(k, ".name") {
								gotNames = append(gotNames, v)
								if v == expectedName {
									found++
								}
							}
							if strings.HasSuffix(k, ".signal") {
								gotSignals = append(gotSignals, v)
							}
							if strings.HasSuffix(k, ".action") {
								gotActions = append(gotActions, v)
							}
						}

						if found == 0 {
							return fmt.Errorf("expected threshold with name %s, got names: %v, all attributes: %v", expectedName, gotNames, a)
						}

						// Validate we have the expected signal
						if !slices.Contains(gotSignals, "SQLI") {
							return fmt.Errorf("expected signal 'SQLI', got signals: %v", gotSignals)
						}

						// Validate we have the expected action
						if !slices.Contains(gotActions, "log") {
							return fmt.Errorf("expected action 'log', got actions: %v", gotActions)
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
