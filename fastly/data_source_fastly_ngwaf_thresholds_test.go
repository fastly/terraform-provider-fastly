package fastly

import (
	"fmt"
	"slices"
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

						expectedName := fmt.Sprintf("test-threshold-%s", h)
						var (
							found      int
							gotNames   []string
							gotSignals []string
							gotActions []string
						)

						// Check for threshold attributes like:
						// "thresholds.0.name": "test-threshold-abc123"
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
							return fmt.Errorf("expected threshold with name %s, got names: %v", expectedName, gotNames)
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
	tf := `
resource "fastly_ngwaf_workspace" "example" {
  name = "tf_%s"
  description = "Test NGWAF Workspace %s"
  mode = "block"
  ip_anonymization = "hashed"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_thresholds" "test_thresholds_workspace" {
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
