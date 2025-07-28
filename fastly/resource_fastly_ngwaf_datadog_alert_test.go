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
	datadogAlertDescription = "Test NGWAF Datadog Alert"
	datadogAlertKey         = "123456789"
	datadogAlertKeyUpdated  = "987654321"
	datadogAlertSite        = "us1"
	datadogAlertSiteUpdated = "us2"
)

func TestAccFastlyNGWAFDatadogAlert_validate(t *testing.T) {
	var (
		datadogAlertID string
		workspaceID    string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFDatadogAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFDatadogAlertConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "description", datadogAlertDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "integration_key", datadogAlertKey),
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "integration_site", datadogAlertSite),
					testAccNGWAFDatadogAlertExists("fastly_ngwaf_datadog_alert.test_datadog_alert", "fastly_ngwaf_workspace.test_datadog_alert_workspace", &datadogAlertID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFDatadogAlertConfigUpdate(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "description", datadogAlertDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "integration_key", datadogAlertKeyUpdated),
					resource.TestCheckResourceAttr("fastly_ngwaf_datadog_alert.test_datadog_alert", "integration_site", datadogAlertSiteUpdated),
					testAccNGWAFDatadogAlertExists("fastly_ngwaf_datadog_alert.test_datadog_alert", "fastly_ngwaf_workspace.test_datadog_alert_workspace", &datadogAlertID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_datadog_alert.test_datadog_alert",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF datadog alert ID: %s", datadogAlertID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, datadogAlertID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFDatadogAlertExists(datadogAlertName, workspaceName string, datadogAlertID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datadogAlertName]
		if !ok {
			return fmt.Errorf("Not found: %s", datadogAlertName)
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

		*datadogAlertID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF datadog alert ID: %s", *datadogAlertID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFDatadogAlertDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_datadog_alert" {
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

func testAccNGWAFDatadogAlertConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_datadog_alert_workspace" {
  name                         = "%s"
  description                  = "Test NGWAF Workspace"
  mode                         = "block"
  ip_anonymization            = "hashed"
  client_ip_headers           = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_datadog_alert" "test_datadog_alert" {
  description                  = "%s"
  integration_key              = "%s"
  integration_site             = "%s"
  workspace_id                 = fastly_ngwaf_workspace.test_datadog_alert_workspace.id
}
`, workspaceName, datadogAlertDescription, datadogAlertKey, datadogAlertSite)
}

func testAccNGWAFDatadogAlertConfigUpdate(workspaceName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_datadog_alert_workspace" {
  name                         = "%s"
  description                  = "Test NGWAF Workspace"
  mode                         = "block"
  ip_anonymization            = "hashed"
  client_ip_headers           = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_datadog_alert" "test_datadog_alert" {
  description                  = "%s"
  integration_key              = "%s"
  integration_site             = "%s"
  workspace_id                 = fastly_ngwaf_workspace.test_datadog_alert_workspace.id
}
`, workspaceName, datadogAlertDescription, datadogAlertKeyUpdated, datadogAlertSiteUpdated)
}
