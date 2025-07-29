package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	ddalerts "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

const (
	AlertDatadogIntegrationDescription = "Test NGWAF Datadog Alert"
	AlertDatadogIntegrationKey         = "123456789"
	AlertDatadogIntegrationKeyUpdated  = "987654321"
	AlertDatadogIntegrationSite        = "us1"
	AlertDatadogIntegrationSiteUpdated = "us3"
)

func TestAccFastlyNGWAFAlertDatadogIntegration_validate(t *testing.T) {
	var (
		AlertDatadogIntegrationID string
		workspaceID               string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertDatadogIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertDatadogIntegrationConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "description", AlertDatadogIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "key", AlertDatadogIntegrationKey),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "site", AlertDatadogIntegrationSite),
					testAccNGWAFAlertDatadogIntegrationExists("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "fastly_ngwaf_workspace.test_datadog_alert_workspace", &AlertDatadogIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertDatadogIntegrationConfigUpdate(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "description", AlertDatadogIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "key", AlertDatadogIntegrationKeyUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "site", AlertDatadogIntegrationSiteUpdated),
					testAccNGWAFAlertDatadogIntegrationExists("fastly_ngwaf_alert_datadog_integration.test_datadog_alert", "fastly_ngwaf_workspace.test_datadog_alert_workspace", &AlertDatadogIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_datadog_integration.test_datadog_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF datadog alert ID: %s", AlertDatadogIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, AlertDatadogIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertDatadogIntegrationExists(AlertDatadogIntegrationName, workspaceName string, AlertDatadogIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[AlertDatadogIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", AlertDatadogIntegrationName)
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
			return fmt.Errorf("Unable to retrieve NGWAF Datadog Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Datadog Alert %s not found in API", wID)
		}

		*AlertDatadogIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF datadog alert ID: %s", *AlertDatadogIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertDatadogIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_datadog_integration" {
			continue
		}

		_, err := ddalerts.Get(context.TODO(), conn, &ddalerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Datadog Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertDatadogIntegrationConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_datadog_alert_workspace" {
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

resource "fastly_ngwaf_alert_datadog_integration" "test_datadog_alert" {
  description      = "%s"
  key              = "%s"
  site             = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_datadog_alert_workspace.id
}
`, workspaceName, AlertDatadogIntegrationDescription, AlertDatadogIntegrationKey, AlertDatadogIntegrationSite)
}

func testAccNGWAFAlertDatadogIntegrationConfigUpdate(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_datadog_alert_workspace" {
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

resource "fastly_ngwaf_alert_datadog_integration" "test_datadog_alert" {
  description      = "%s"
  key              = "%s"
  site             = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_datadog_alert_workspace.id
}
`, workspaceName, AlertDatadogIntegrationDescription, AlertDatadogIntegrationKeyUpdated, AlertDatadogIntegrationSiteUpdated)
}
