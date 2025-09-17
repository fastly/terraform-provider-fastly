package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pagerdutyAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
)

const (
	AlertPagerDutyIntegrationDescription = "Test NGWAF PagerDuty Alert"
	AlertPagerDutyIntegrationKey         = "1234567890abcdef1234567890abcdef"
	AlertPagerDutyIntegrationKeyUpdated  = "abcdef1234567890abcdef1234567890"
)

func TestAccFastlyNGWAFAlertPagerDutyIntegration_validate(t *testing.T) {
	var (
		alertPagerDutyIntegrationID string
		workspaceID                 string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertPagerDutyIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertPagerDutyIntegrationConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "description", AlertPagerDutyIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "key", AlertPagerDutyIntegrationKey),
					testAccNGWAFAlertPagerDutyIntegrationExists("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "fastly_ngwaf_workspace.test_pagerduty_alert_workspace", &alertPagerDutyIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertPagerDutyIntegrationConfigUpdate(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "description", AlertPagerDutyIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "key", AlertPagerDutyIntegrationKeyUpdated),
					testAccNGWAFAlertPagerDutyIntegrationExists("fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert", "fastly_ngwaf_workspace.test_pagerduty_alert_workspace", &alertPagerDutyIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_pagerduty_integration.test_pagerduty_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF pagerduty alert ID: %s", alertPagerDutyIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertPagerDutyIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertPagerDutyIntegrationExists(alertPagerDutyIntegrationName, workspaceName string, alertPagerDutyIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertPagerDutyIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertPagerDutyIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := pagerdutyAlerts.Get(context.TODO(), conn, &pagerdutyAlerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF PagerDuty Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF PagerDuty Alert %s not found in API", wID)
		}

		*alertPagerDutyIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF pagerduty alert ID: %s", *alertPagerDutyIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertPagerDutyIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_pagerduty_integration" {
			continue
		}

		_, err := pagerdutyAlerts.Get(context.TODO(), conn, &pagerdutyAlerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF PagerDuty Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertPagerDutyIntegrationConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_pagerduty_alert_workspace" {
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

resource "fastly_ngwaf_alert_pagerduty_integration" "test_pagerduty_alert" {
  description      = "%s"
  key              = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_pagerduty_alert_workspace.id
}
`, workspaceName, AlertPagerDutyIntegrationDescription, AlertPagerDutyIntegrationKey)
}

func testAccNGWAFAlertPagerDutyIntegrationConfigUpdate(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_pagerduty_alert_workspace" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_alert_pagerduty_integration" "test_pagerduty_alert" {
  description      = "%s"
  key              = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_pagerduty_alert_workspace.id
}
`, workspaceName, AlertPagerDutyIntegrationDescription, AlertPagerDutyIntegrationKeyUpdated)
}
