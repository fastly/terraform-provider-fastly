package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFAlertMailingListIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertMailingListIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_mailing_list_integration.example"]
						a := r.Primary.Attributes
						mailingListAlerts, err := strconv.Atoi(a["mailing_list_alerts.#"])
						if err != nil {
							return err
						}

						if mailingListAlerts != 1 {
							return fmt.Errorf("want: %v, got: %v", h, mailingListAlerts)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertMailingListIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_mailing_list_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_mailing_list_integration" "example_1" {
  address          = "test@fastly.com"
  description      = "some integration"
  workspace_id     = fastly_ngwaf_workspace.test_mailing_list_alerts_workspace.id
}

data "fastly_ngwaf_alert_mailing_list_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_mailing_list_alerts_workspace,
    fastly_ngwaf_alert_mailing_list_integration.example_1
  ]
  workspace_id = fastly_ngwaf_workspace.test_mailing_list_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h)
}
