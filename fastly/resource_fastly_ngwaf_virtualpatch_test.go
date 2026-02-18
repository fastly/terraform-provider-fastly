package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	ws "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/virtualpatches"
)

func TestAccFastlyNGWAFVirtualPatch_validate(t *testing.T) {
	newWorkspaceName := fmt.Sprintf("Test Virtual Patch WS %s", acctest.RandString(10))
	virtualPatchID := "CVE-2017-5638"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFVirtualPatchConfig(newWorkspaceName, virtualPatchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "action", "block"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "enabled", "true"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_virtual_patches.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "virtual_patch_id", virtualPatchID),
					testAccNGWAFVirtualPatchExists("fastly_ngwaf_virtual_patches.sample"),
				),
			},
			{
				Config: testAccNGWAFVirtualPatchConfigUpdate(newWorkspaceName, virtualPatchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "action", "log"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "enabled", "false"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_virtual_patches.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "virtual_patch_id", virtualPatchID),
					testAccNGWAFVirtualPatchExists("fastly_ngwaf_virtual_patches.sample"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_ngwaf_virtual_patches.sample",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNGWAFVirtualPatchImportID("fastly_ngwaf_virtual_patches.sample"),
			},
		},
	})
}

func testAccNGWAFVirtualPatchImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["virtual_patch_id"]), nil
	}
}

func testAccNGWAFVirtualPatchExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		virtualpatch, err := ws.Get(context.TODO(), conn, &ws.GetInput{
			WorkspaceID:    gofastly.ToPointer(rs.Primary.Attributes["workspace_id"]),
			VirtualPatchID: gofastly.ToPointer(rs.Primary.Attributes["virtual_patch_id"]),
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Virtual Patch %s: %v", rs.Primary.ID, err)
		}
		if virtualpatch == nil {
			return fmt.Errorf("NGWAF Virtual Patch %s not found in API", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNGWAFVirtualPatchConfig(workspaceName, virtualPatchID string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                         = "%s"
  description                  = "Test VP Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

  resource "fastly_ngwaf_virtual_patches" "sample" {
    action            = "block"
    enabled           = true
    virtual_patch_id  = "%s"
    workspace_id      = fastly_ngwaf_workspace.example.id
  }
  `, workspaceName, virtualPatchID)
}

func testAccNGWAFVirtualPatchConfigUpdate(workspaceName, virtualPatchID string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                         = "%s"
  description                  = "Test VP Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

  resource "fastly_ngwaf_virtual_patches" "sample" {
    action            = "log"
    enabled           = false
    virtual_patch_id  = "%s"
    workspace_id      = fastly_ngwaf_workspace.example.id
  }
  `, workspaceName, virtualPatchID)
}
