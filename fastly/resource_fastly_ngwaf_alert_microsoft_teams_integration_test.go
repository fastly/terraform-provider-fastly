package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	microsoftTeamsAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
)

const (
	AlertMicrosoftTeamsIntegrationDescription = "Test NGWAF Microsoft Teams Alert"
	AlertMicrosoftTeamsIntegrationURL         = "https://example.com/microsoft-teams/my-service"
	AlertMicrosoftTeamsIntegrationURLUpdated  = "https://example.com/microsoft-teams/my-service-2"
)

func TestAccFastlyNGWAFAlertMicrosoftTeamsIntegration_validate(t *testing.T) {
	var (
		alertMicrosoftTeamsIntegrationID string
		workspaceID                      string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertMicrosoftTeamsIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertMicrosoftTeamsIntegrationConfig(workspaceName, AlertMicrosoftTeamsIntegrationDescription, AlertMicrosoftTeamsIntegrationURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "description", AlertMicrosoftTeamsIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "webhook", AlertMicrosoftTeamsIntegrationURL),
					testAccNGWAFAlertMicrosoftTeamsIntegrationExists("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "fastly_ngwaf_workspace.test_microsoft_teams_alert_workspace", &alertMicrosoftTeamsIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertMicrosoftTeamsIntegrationConfig(workspaceName, AlertMicrosoftTeamsIntegrationDescription, AlertMicrosoftTeamsIntegrationURLUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "description", AlertMicrosoftTeamsIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "webhook", AlertMicrosoftTeamsIntegrationURLUpdated),
					testAccNGWAFAlertMicrosoftTeamsIntegrationExists("fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert", "fastly_ngwaf_workspace.test_microsoft_teams_alert_workspace", &alertMicrosoftTeamsIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_microsoft_teams_integration.test_microsoft_teams_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF Microsoft Teams alert ID: %s", alertMicrosoftTeamsIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertMicrosoftTeamsIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertMicrosoftTeamsIntegrationExists(alertMicrosoftTeamsIntegrationName, workspaceName string, alertMicrosoftTeamsIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertMicrosoftTeamsIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertMicrosoftTeamsIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := microsoftTeamsAlerts.Get(context.TODO(), conn, &microsoftTeamsAlerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Microsoft Teams Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Microsoft Teams Alert %s not found in API", wID)
		}

		*alertMicrosoftTeamsIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF Microsoft Teams alert ID: %s", *alertMicrosoftTeamsIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertMicrosoftTeamsIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_microsoft_teams_integration" {
			continue
		}

		_, err := microsoftTeamsAlerts.Get(context.TODO(), conn, &microsoftTeamsAlerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF MicrosoftTeams Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertMicrosoftTeamsIntegrationConfig(workspaceName, description, webhook string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_microsoft_teams_alert_workspace" {
  name                           = "%s"
  description                    = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_alert_microsoft_teams_integration" "test_microsoft_teams_alert" {
  description      = "%s"
  webhook          = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_microsoft_teams_alert_workspace.id
}
`, workspaceName, description, webhook)
}
