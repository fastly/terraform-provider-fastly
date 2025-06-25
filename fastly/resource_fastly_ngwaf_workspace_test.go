package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	ws "github.com/fastly/go-fastly/v10/fastly/ngwaf/v1/workspaces"
)

func TestAccFastlyNGWAFWorkspace_validate(t *testing.T) {
	workspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))
	newWorkspaceName := fmt.Sprintf("NGWAF Workspace %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "description", "Test NGWAF Workspace"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "mode", "block"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "ip_anonymization", "hashed"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "client_ip_headers.0", "X-Forwarded-For"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "client_ip_headers.1", "X-Real-IP"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "default_blocking_response_code", "429"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.one_minute", "100"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.ten_minutes", "500"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.one_hour", "1000"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.immediate", "true"),
					testAccNGWAFWorkspaceExists("fastly_ngwaf_workspace.example"),
				),
			},
			{
				Config: testAccNGWAFWorkspaceConfigUpdate(newWorkspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", newWorkspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "description", "Test NGWAF Workspace Updated"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "mode", "log"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "ip_anonymization", "hashed"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "client_ip_headers.0", "True-Client-IP"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "default_blocking_response_code", "406"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.one_minute", "200"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.ten_minutes", "1000"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.one_hour", "2000"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "attack_signal_thresholds.0.immediate", "false"),
					testAccNGWAFWorkspaceExists("fastly_ngwaf_workspace.example"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNGWAFWorkspaceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		workspace, err := ws.Get(conn, &ws.GetInput{
			WorkspaceID: &rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Workspace %s: %v", rs.Primary.ID, err)
		}

		if workspace == nil {
			return fmt.Errorf("NGWAF Workspace %s not found in API", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNGWAFWorkspaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_workspace" {
			continue
		}

		_, err := ws.Get(conn, &ws.GetInput{
			WorkspaceID: &rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("NGWAF Workspace %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccNGWAFWorkspaceConfig(name string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
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
`, name)
}

func testAccNGWAFWorkspaceConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                         = "%s"
  description                  = "Test NGWAF Workspace Updated"
  mode                         = "log"
  ip_anonymization            = "hashed"
  client_ip_headers           = ["True-Client-IP"]
  default_blocking_response_code = 406

  attack_signal_thresholds {
    one_minute  = 200
    ten_minutes = 1000
    one_hour    = 2000
    immediate   = false
  }
}
`, name)
}
