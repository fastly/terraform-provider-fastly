package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFWorkspaceList_basic(t *testing.T) {
	workspaceName := fmt.Sprintf("tf-ws-%s", acctest.RandString(5))
	listName := fmt.Sprintf("ws-list-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil, // Lists are deleted when workspace is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceListConfig(workspaceName, listName, "Initial workspace list", []string{"10.0.0.1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "name", listName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "type", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "description", "Initial workspace list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "entries.0", "10.0.0.1"),
				),
			},
			{
				Config: testAccNGWAFWorkspaceListConfig(workspaceName, listName, "Updated workspace list", []string{"192.168.1.1", "172.16.0.1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "description", "Updated workspace list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "entries.0", "192.168.1.1"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_list.example", "entries.1", "172.16.0.1"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace_list.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					workspace := s.RootModule().Resources["fastly_ngwaf_workspace.example"]
					list := s.RootModule().Resources["fastly_ngwaf_workspace_list.example"]
					return fmt.Sprintf("%s/%s", workspace.Primary.ID, list.Primary.ID), nil
				},
			},
		},
	})
}

func testAccNGWAFWorkspaceListConfig(workspaceName, listName, description string, entries []string) string {
	entriesFormatted := ""
	for _, e := range entries {
		entriesFormatted += fmt.Sprintf("\"%s\", ", e)
	}
	entriesFormatted = entriesFormatted[:len(entriesFormatted)-2] // trim trailing ", "

	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Workspace for list testing"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Real-IP"]
  default_blocking_response_code = 403

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_list" "example" {
  workspace_id = fastly_ngwaf_workspace.example.id
  name         = "%s"
  description  = "%s"
  type         = "ip"
  entries      = [%s]
}
`, workspaceName, listName, description, entriesFormatted)
}
