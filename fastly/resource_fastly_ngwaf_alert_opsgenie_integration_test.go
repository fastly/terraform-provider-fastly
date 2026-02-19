package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	opsgenieAlerts "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/opsgenie"
)

const (
	AlertOpsgenieIntegrationDescription = "Test NGWAF Opsgenie Alert"
	AlertOpsgenieIntegrationKey         = "123456789"
	AlertOpsgenieIntegrationKeyUpdated  = "987654321"
)

func TestAccFastlyNGWAFAlertOpsgenieIntegration_validate(t *testing.T) {
	var (
		alertOpsgenieIntegrationID string
		workspaceID                string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertOpsgenieIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertOpsgenieIntegrationConfig(workspaceName, AlertOpsgenieIntegrationDescription, AlertOpsgenieIntegrationKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "description", AlertOpsgenieIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "key", AlertOpsgenieIntegrationKey),
					testAccNGWAFAlertOpsgenieIntegrationExists("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "fastly_ngwaf_workspace.test_opsgenie_alert_workspace", &alertOpsgenieIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertOpsgenieIntegrationConfig(workspaceName, AlertOpsgenieIntegrationDescription, AlertOpsgenieIntegrationKeyUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "description", AlertOpsgenieIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "key", AlertOpsgenieIntegrationKeyUpdated),
					testAccNGWAFAlertOpsgenieIntegrationExists("fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert", "fastly_ngwaf_workspace.test_opsgenie_alert_workspace", &alertOpsgenieIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_opsgenie_integration.test_opsgenie_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF opsgenie alert ID: %s", alertOpsgenieIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertOpsgenieIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertOpsgenieIntegrationExists(alertOpsgenieIntegrationName, workspaceName string, alertOpsgenieIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertOpsgenieIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertOpsgenieIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := opsgenieAlerts.Get(context.TODO(), conn, &opsgenieAlerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Opsgenie Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Opsgenie Alert %s not found in API", wID)
		}

		*alertOpsgenieIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF opsgenie alert ID: %s", *alertOpsgenieIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertOpsgenieIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_opsgenie_integration" {
			continue
		}

		_, err := opsgenieAlerts.Get(context.TODO(), conn, &opsgenieAlerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Opsgenie Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertOpsgenieIntegrationConfig(workspaceName, description, key string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_opsgenie_alert_workspace" {
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

resource "fastly_ngwaf_alert_opsgenie_integration" "test_opsgenie_alert" {
  description      = "%s"
  key              = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_opsgenie_alert_workspace.id
}
`, workspaceName, description, key)
}
