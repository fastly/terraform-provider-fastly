package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	webhookAlerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/webhook"
)

const (
	AlertWebhookIntegrationDescription = "Test NGWAF Webhook Alert"
	AlertWebhookIntegrationURL         = "https://example.com/webhooks/my-service"
	AlertWebhookIntegrationURLUpdated  = "https://example.com/webhooks/my-service-2"
)

func TestAccFastlyNGWAFAlertWebhookIntegration_validate(t *testing.T) {
	var (
		alertWebhookIntegrationID string
		workspaceID               string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertWebhookIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertWebhookIntegrationConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "description", AlertWebhookIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "webhook", AlertWebhookIntegrationURL),
					testAccNGWAFAlertWebhookIntegrationExists("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "fastly_ngwaf_workspace.test_webhook_alert_workspace", &alertWebhookIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertWebhookIntegrationConfigUpdate(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "description", AlertWebhookIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "webhook", AlertWebhookIntegrationURLUpdated),
					testAccNGWAFAlertWebhookIntegrationExists("fastly_ngwaf_alert_webhook_integration.test_webhook_alert", "fastly_ngwaf_workspace.test_webhook_alert_workspace", &alertWebhookIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_webhook_integration.test_webhook_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF webhook alert ID: %s", alertWebhookIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertWebhookIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertWebhookIntegrationExists(alertWebhookIntegrationName, workspaceName string, alertWebhookIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertWebhookIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertWebhookIntegrationName)
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
			return fmt.Errorf("Unable to retrieve NGWAF Webhook Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Webhook Alert %s not found in API", wID)
		}

		*alertWebhookIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF webhook alert ID: %s", *alertWebhookIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertWebhookIntegrationDestroy(s *terraform.State) error {
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
			return fmt.Errorf("NGWAF Webhook Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertWebhookIntegrationConfig(workspaceName string) string {
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
`, workspaceName, AlertWebhookIntegrationDescription, AlertWebhookIntegrationURL)
}

func testAccNGWAFAlertWebhookIntegrationConfigUpdate(workspaceName string) string {
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
`, workspaceName, AlertWebhookIntegrationDescription, AlertWebhookIntegrationURLUpdated)
}
