package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wsr "github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/redactions"
)

func TestAccFastlyNGWAFRedaction_validate(t *testing.T) {
	var (
		redactionID string
		workspaceID string
	)
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))
	redactionField := fmt.Sprintf("redaction field %s", acctest.RandString(10))
	newRedactionField := fmt.Sprintf("redaction field %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFRedactionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFRedactionConfig(workspaceName, redactionField),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_redaction.test_redaction", "field", redactionField),
					resource.TestCheckResourceAttr("fastly_ngwaf_redaction.test_redaction", "type", "request_parameter"),
					testAccNGWAFRedactionExists("fastly_ngwaf_redaction.test_redaction", "fastly_ngwaf_workspace.test_redactions_workspace", &redactionID, &workspaceID),
				),
			},
			{
				Config: testAccNGWAFRedactionConfigUpdate(workspaceName, newRedactionField),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_redaction.test_redaction", "field", newRedactionField),
					resource.TestCheckResourceAttr("fastly_ngwaf_redaction.test_redaction", "type", "request_header"),
					testAccNGWAFRedactionExists("fastly_ngwaf_redaction.test_redaction", "fastly_ngwaf_workspace.test_redactions_workspace", &redactionID, &workspaceID),
				),
			},
			{
				ResourceName: "fastly_ngwaf_redaction.test_redaction",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					log.Printf("[DEBUG] IMPORT TEST: NGWAF redaction ID: %s", redactionID)
					log.Printf("[DEBUG] IMPORT TEST: NGWAF workspace ID: %s", workspaceID)
					return fmt.Sprintf("%s/%s", workspaceID, redactionID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFRedactionExists(redactionName, workspaceName string, redactionID, workspaceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[redactionName]
		if !ok {
			return fmt.Errorf("Not found: %s", redactionName)
		}
		ws, ok := s.RootModule().Resources[workspaceName]
		if !ok {
			return fmt.Errorf("Not found: %s", workspaceName)
		}
		rID := rs.Primary.ID
		wID := ws.Primary.ID

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := wsr.Get(context.TODO(), conn, &wsr.GetInput{
			RedactionID: &rID,
			WorkspaceID: &wID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Redaction %s: %v", rID, err)
		}

		if latest == nil {
			return fmt.Errorf("NGWAF Redaction %s not found in API", wID)
		}

		*redactionID = rID
		*workspaceID = wID
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF redaction ID: %s", *redactionID)
		log.Printf("[DEBUG] EXISTS IMPORT: NGWAF workspace ID: %s", *workspaceID)

		return nil
	}
}

func testAccCheckNGWAFRedactionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_redaction" {
			continue
		}

		_, err := wsr.Get(context.TODO(), conn, &wsr.GetInput{
			RedactionID: &rs.Primary.ID,
			WorkspaceID: &wsID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Redaction %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFRedactionConfig(workspaceName, redactionField string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_redactions_workspace" {
  name                         = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_redaction" "test_redaction" {
  field                        = "%s"
  type                         = "request_parameter"
  workspace_id                 = fastly_ngwaf_workspace.test_redactions_workspace.id
}
`, workspaceName, redactionField)
}

func testAccNGWAFRedactionConfigUpdate(workspaceName, redactionField string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "test_redactions_workspace" {
  name                         = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_redaction" "test_redaction" {
  field                        = "%s"
  type                         = "request_header"
  workspace_id                 = fastly_ngwaf_workspace.test_redactions_workspace.id
}
`, workspaceName, redactionField)
}
