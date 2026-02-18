package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	mailingListAlerts "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/mailinglist"
)

const (
	AlertMailingListIntegrationDescription    = "Test NGWAF MailingList Alert"
	AlertMailingListIntegrationAddress        = "test@fastly.com"
	AlertMailingListIntegrationAddressUpdated = "test2@fastly.com"
)

func TestAccFastlyNGWAFAlertMailingListIntegration_validate(t *testing.T) {
	var (
		alertMailingListIntegrationID string
		workspaceID                   string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAlertMailingListIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAlertMailingListIntegrationConfig(workspaceName, AlertMailingListIntegrationDescription, AlertMailingListIntegrationAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "description", AlertMailingListIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "address", AlertMailingListIntegrationAddress),
					testAccNGWAFAlertMailingListIntegrationExists("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "fastly_ngwaf_workspace.test_mailing_list_alert_workspace", &alertMailingListIntegrationID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFAlertMailingListIntegrationConfig(workspaceName, AlertMailingListIntegrationDescription, AlertMailingListIntegrationAddressUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "description", AlertMailingListIntegrationDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "address", AlertMailingListIntegrationAddressUpdated),
					testAccNGWAFAlertMailingListIntegrationExists("fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert", "fastly_ngwaf_workspace.test_mailing_list_alert_workspace", &alertMailingListIntegrationID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_alert_mailing_list_integration.test_mailing_list_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF mailingList alert ID: %s", alertMailingListIntegrationID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, alertMailingListIntegrationID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFAlertMailingListIntegrationExists(alertMailingListIntegrationName, workspaceName string, alertMailingListIntegrationID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[alertMailingListIntegrationName]
		if !ok {
			return fmt.Errorf("Not found: %s", alertMailingListIntegrationName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := mailingListAlerts.Get(context.TODO(), conn, &mailingListAlerts.GetInput{
			AlertID:     &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Mailing List Alert %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Mailing List Alert %s not found in API", wID)
		}

		*alertMailingListIntegrationID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF mailing list alert ID: %s", *alertMailingListIntegrationID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFAlertMailingListIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_alert_mailing_list_integration" {
			continue
		}

		_, err := mailingListAlerts.Get(context.TODO(), conn, &mailingListAlerts.GetInput{
			AlertID:     &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Mailing List Alert %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFAlertMailingListIntegrationConfig(workspaceName, description, address string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_mailing_list_alert_workspace" {
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

resource "fastly_ngwaf_alert_mailing_list_integration" "test_mailing_list_alert" {
  address          = "%s"
  description      = "%s"
  workspace_id     = fastly_ngwaf_workspace.test_mailing_list_alert_workspace.id
}
`, workspaceName, address, description)
}
