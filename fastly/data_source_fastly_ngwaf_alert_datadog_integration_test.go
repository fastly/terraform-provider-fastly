package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFAlertDatadogIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertDatadogIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_datadog_integration.example"]
						a := r.Primary.Attributes
						datadogAlerts, err := strconv.Atoi(a["datadog_alerts.#"])
						if err != nil {
							return err
						}

						if datadogAlerts != 1 {
							return fmt.Errorf("want: %v, got: %v", h, datadogAlerts)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertDatadogIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_datadog_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_datadog_integration" "example_1" {
  description      = "%s 1"
  key              = "123456789"
  site             = "us1"
  workspace_id     = fastly_ngwaf_workspace.test_datadog_alerts_workspace.id
}

data "fastly_ngwaf_alert_datadog_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_datadog_alerts_workspace,
    fastly_ngwaf_alert_datadog_integration.example_1
  ]
  workspace_id = fastly_ngwaf_workspace.test_datadog_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
