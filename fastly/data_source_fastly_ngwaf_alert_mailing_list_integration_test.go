package fastly

import (
	"fmt"
	"strings"
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
						var (
							found int
							got   []string
						)
						for k, v := range a {
							if strings.HasSuffix(k, ".description") {
								got = append(got, v)
								if strings.Contains(v, h) {
									found++
								}
							}
						}

						if found != 1 {
							return fmt.Errorf("want: %v, got: %v", h, got)
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

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_alert_mailing_list_integration" "example_1" {
  address          = "test@fastly.com"
  description      = "%s 1"
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
	return fmt.Sprintf(tf, h, h)
}
