package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	ddalerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/jira"
)

const (
	AlertJiraIntegrationDescription      = "Test NGWAF Jira Alert"
	AlertJiraIntegrationHost             = "https://mycompany.atlassian.net"
	AlertJiraIntegrationHostUpdated      = "https://mycompany1.atlassian.net"
	AlertJiraIntegrationIssueType        = "task"
	AlertJiraIntegrationIssueTypeUpdated = "bug"
	AlertJiraIntegrationKey              = "a1b2c3d4e5f6789012345678901234567"
	AlertJiraIntegrationKeyUpdated       = "a1b2c3d4e5f6789012345678901234568"
	AlertJiraIntegrationProject          = "test"
	AlertJiraIntegrationProjectUpdated   = "test2"
	AlertJiraIntegrationUsername         = "user"
	AlertJiraIntegrationUsernameUpdated  = "user2"
)

func TestAccFastlyNGWAFAlertJiraIntegration_validate(t *testing.T) {
	var (
		alertJiraIntegrationID string
		workspaceID            string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertJiraIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertJiraIntegrationConfig(workspaceName, AlertJiraIntegrationDescription, AlertJiraIntegrationHost, AlertJiraIntegrationIssueType, AlertJiraIntegrationKey, AlertJiraIntegrationProject, AlertJiraIntegrationUsername),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "description", AlertJiraIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "host", AlertJiraIntegrationHost),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "issue_type", AlertJiraIntegrationIssueType),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "key", AlertJiraIntegrationKey),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "project", AlertJiraIntegrationProject),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "username", AlertJiraIntegrationUsername),
					testAccNGWAFAlertJiraIntegrationExists("fastly_ngwaf_alert_jira_integration.test_jira_alert", "fastly_ngwaf_workspace.test_jira_alert_workspace", &alertJiraIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertJiraIntegrationConfig(workspaceName, AlertJiraIntegrationDescription, AlertJiraIntegrationHostUpdated, AlertJiraIntegrationIssueTypeUpdated, AlertJiraIntegrationKeyUpdated, AlertJiraIntegrationProjectUpdated, AlertJiraIntegrationUsernameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "description", AlertJiraIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "host", AlertJiraIntegrationHostUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "issue_type", AlertJiraIntegrationIssueTypeUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "key", AlertJiraIntegrationKeyUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "project", AlertJiraIntegrationProjectUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_jira_integration.test_jira_alert", "username", AlertJiraIntegrationUsernameUpdated),
					testAccNGWAFAlertJiraIntegrationExists("fastly_ngwaf_alert_jira_integration.test_jira_alert", "fastly_ngwaf_workspace.test_jira_alert_workspace", &alertJiraIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_jira_integration.test_jira_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF jira alert ID: %s", alertJiraIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertJiraIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertJiraIntegrationExists(alertJiraIntegrationName, workspaceName string, alertJiraIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertJiraIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertJiraIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := ddalerts.Get(context.TODO(), conn, &ddalerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Jira Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Jira Alert %s not found in API", wID)
		}

		*alertJiraIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF jira alert ID: %s", *alertJiraIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertJiraIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_jira_integration" {
			continue
		}

		_, err := ddalerts.Get(context.TODO(), conn, &ddalerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Jira Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertJiraIntegrationConfig(workspaceName, description, host, issueType, key, project, username string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_jira_alert_workspace" {
  name                           = "%s"
  description                    = "Test NGWAF Workspace"
  mode                           = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_alert_jira_integration" "test_jira_alert" {
  description      = "%s"
  host             = "%s"
  issue_type       = "%s"
  key              = "%s"
  project          = "%s"
  username         = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_jira_alert_workspace.id
}
`, workspaceName, description, host, issueType, key, project, username)
}
