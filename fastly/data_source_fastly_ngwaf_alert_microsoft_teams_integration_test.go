package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAlertMicrosoftTeamsIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertMicrosoftTeamsIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_microsoft_teams_integration.example"]
						a := r.Primary.Attributes

						// Check that we have at least one webhook alert with an ID
						microsoftTeamsCount := a["microsoft_teams_alerts.#"]
						if microsoftTeamsCount == "0" {
							return fmt.Errorf("expected at least one microsoft teams alert, got %s", microsoftTeamsCount)
						}

						// Check that the first webhook alert has an ID
						microsoftTeamsID := a["microsoft_teams_alerts.0.id"]
						if microsoftTeamsID == "" {
							return fmt.Errorf("expected microsoft teams alert to have an ID, got empty string")
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertMicrosoftTeamsIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_microsoft_teams_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_microsoft_teams_integration" "sample" {
  description      = "%s 1"
  webhook          = "https://example.com/microsoft-teams/my-service"
  workspace_id     = fastly_ngwaf_workspace.test_microsoft_teams_alerts_workspace.id
}

data "fastly_ngwaf_alert_microsoft_teams_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_microsoft_teams_alerts_workspace,
    fastly_ngwaf_alert_microsoft_teams_integration.sample
  ]
  workspace_id = fastly_ngwaf_workspace.test_microsoft_teams_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
