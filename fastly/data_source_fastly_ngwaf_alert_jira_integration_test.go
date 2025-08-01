package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFAlertJiraIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertJiraIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_jira_integration.example"]
						a := r.Primary.Attributes
						jiraAlerts, err := strconv.Atoi(a["jira_alerts.#"])
						if err != nil {
							return err
						}

						if jiraAlerts != 1 {
							return fmt.Errorf("want: %v, got: %v", h, jiraAlerts)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertJiraIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_jira_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_jira_integration" "example_1" {
  description      = "%s 1"
  host             = "https://mycompany.atlassian.net"
  issue_type       = "task"
  key              = "a1b2c3d4e5f6789012345678901234567"
  project          = "test"
  username         = "user"
  workspace_id     = fastly_ngwaf_workspace.test_jira_alerts_workspace.id
}

data "fastly_ngwaf_alert_jira_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_jira_alerts_workspace,
    fastly_ngwaf_alert_jira_integration.example_1
  ]
  workspace_id = fastly_ngwaf_workspace.test_jira_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
