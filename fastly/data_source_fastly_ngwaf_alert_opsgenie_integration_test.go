package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFAlertOpsgenieIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertOpsgenieIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_opsgenie_integration.example"]
						a := r.Primary.Attributes
						opsgenieAlerts, err := strconv.Atoi(a["opsgenie_alerts.#"])
						if err != nil {
							return err
						}

						if opsgenieAlerts != 1 {
							return fmt.Errorf("want: %v, got: %v", h, opsgenieAlerts)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertOpsgenieIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_opsgenie_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_opsgenie_integration" "example_1" {
  description      = "%s 1"
  key              = "123456789"
  workspace_id     = fastly_ngwaf_workspace.test_opsgenie_alerts_workspace.id
}

data "fastly_ngwaf_alert_opsgenie_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_opsgenie_alerts_workspace,
    fastly_ngwaf_alert_opsgenie_integration.example_1
  ]
  workspace_id = fastly_ngwaf_workspace.test_opsgenie_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
