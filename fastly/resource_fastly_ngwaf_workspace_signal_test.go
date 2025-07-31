package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/signals"
)

func TestFlattenNGWAFSignalResponse(t *testing.T) {
	schemaMap := resourceFastlyNGWAFWorkspaceSignal().Schema
	d := schema.TestResourceDataRaw(t, schemaMap, map[string]any{})

	signal := &signals.Signal{
		Description: "Terraform Signal Unit Test Signal",
		Name:        "Signal Unit Test",
		SignalID:    "example-signal-id",
		Scope: signals.Scope{
			Type:      "workspace",
			AppliesTo: []string{"workspace-123"},
		},
	}

	require.NoError(t, flattenNGWAFSignalResponse(d, signal), "flattenNGWAFSignalResponse should not error")

	// Simple value checks
	require.Equal(t, "Terraform Signal Unit Test Signal", d.Get("description"))
	require.Equal(t, "Signal Unit Test", d.Get("name"))
	require.Equal(t, "workspace-123", d.Get("workspace_id"))
}

func TestAccFastlyNGWAFWorkspaceSignal_basic(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	signalDescription := fmt.Sprintf("Terraform Signal Test %s", acctest.RandString(5))
	signalName := fmt.Sprintf("Signal Test %s", acctest.RandString(5))
	updatedSignalDescription := signalDescription + " updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFWorkspaceSignalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceSignalConfig(workspaceName, signalDescription, signalName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_signal.example", "description", signalDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_signal.example", "name", signalName),
				),
			},
			{
				Config: testAccNGWAFWorkspaceSignalConfig(workspaceName, updatedSignalDescription, signalName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_signal.example", "description", updatedSignalDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_signal.example", "name", signalName),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace_signal.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					signal := s.RootModule().Resources["fastly_ngwaf_workspace_signal.example"]
					workspace := s.RootModule().Resources["fastly_ngwaf_workspace.example"]
					return fmt.Sprintf("%s/%s", workspace.Primary.ID, signal.Primary.ID), nil
				},
			},
		},
	})
}

func testAccCheckNGWAFWorkspaceSignalDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	var wsID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		wsID = rs.Primary.ID
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace_signal" {
			continue
		}

		_, err := signals.Get(context.TODO(), conn, &signals.GetInput{
			SignalID: &rs.Primary.ID,
			Scope: &common.Scope{
				Type:      common.ScopeTypeWorkspace,
				AppliesTo: []string{wsID},
			},
		})
		if err == nil {
			return fmt.Errorf("NGWAF account signal %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccNGWAFWorkspaceSignalConfig(workspaceName, signalDescription, signalName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_signal" "example" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  description      = "%s"
  name             = "%s"
}
`, workspaceName, signalDescription, signalName)
}
