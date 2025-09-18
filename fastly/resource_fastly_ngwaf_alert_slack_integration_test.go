package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	webhookAlerts "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/slack"
)

const (
	AlertSlackIntegrationDescription    = "Test NGWAF Slack Alert"
	AlertSlackIntegrationWebhook        = "https://example.com/webhooks/my-service"
	AlertSlackIntegrationWebhookUpdated = "https://example.com/webhooks/my-service-2"
)

func TestAccFastlyNGWAFAlertSlackIntegration_validate(t *testing.T) {
	var (
		alertSlackIntegrationID string
		workspaceID             string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertSlackIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertSlackIntegrationConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "description", AlertSlackIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "webhook", AlertSlackIntegrationWebhook),
					testAccNGWAFAlertSlackIntegrationExists("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "fastly_ngwaf_workspace.test_webhook_alert_workspace", &alertSlackIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertSlackIntegrationConfigUpdate(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "description", AlertSlackIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "webhook", AlertSlackIntegrationWebhookUpdated),
					testAccNGWAFAlertSlackIntegrationExists("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "fastly_ngwaf_workspace.test_webhook_alert_workspace", &alertSlackIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_webhook_integration.test_webhook_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF slack alert ID: %s", alertSlackIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertSlackIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertSlackIntegrationExists(alertSlackIntegrationName, workspaceName string, alertSlackIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertSlackIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertSlackIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := webhookAlerts.Get(context.TODO(), conn, &webhookAlerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Slack Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Slack Alert %s not found in API", wID)
		}

		*alertSlackIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF slack alert ID: %s", *alertSlackIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertSlackIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_webhook_integration" {
			continue
		}

		_, err := webhookAlerts.Get(context.TODO(), conn, &webhookAlerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Slack Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertSlackIntegrationConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_webhook_alert_workspace" {
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

resource "fastly_ngwaf_alert_webhook_integration" "test_webhook_alert" {
  description      = "%s"
  webhook          = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_webhook_alert_workspace.id
}
`, workspaceName, AlertSlackIntegrationDescription, AlertSlackIntegrationWebhook)
}

func testAccNGWAFAlertSlackIntegrationConfigUpdate(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_webhook_alert_workspace" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_alert_webhook_integration" "test_webhook_alert" {
  description      = "%s"
  webhook          = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_webhook_alert_workspace.id
}
`, workspaceName, AlertSlackIntegrationDescription, AlertSlackIntegrationWebhookUpdated)
}
