package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAlertPagerDutyIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertPagerDutyIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_pagerduty_integration.example"]
						a := r.Primary.Attributes

						// Check that we have at least one pagerduty alert with an ID
						pagerdutyCount := a["pagerduty_alerts.#"]
						if pagerdutyCount == "0" {
							return fmt.Errorf("expected at least one pagerduty alert, got %s", pagerdutyCount)
						}

						// Check that the first pagerduty alert has an ID
						pagerdutyID := a["pagerduty_alerts.0.id"]
						if pagerdutyID == "" {
							return fmt.Errorf("expected pagerduty alert to have an ID, got empty string")
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertPagerDutyIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_pagerduty_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_pagerduty_integration" "sample" {
  description      = "%s 1"
  key              = "1234567890abcdef1234567890abcdef"
  workspace_id     = fastly_ngwaf_workspace.test_pagerduty_alerts_workspace.id
}

data "fastly_ngwaf_alert_pagerduty_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_pagerduty_alerts_workspace,
    fastly_ngwaf_alert_pagerduty_integration.sample
  ]
  workspace_id = fastly_ngwaf_workspace.test_pagerduty_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
